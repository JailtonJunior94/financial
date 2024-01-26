package server

import (
	"fmt"
	"net"
	"net/http"
	"time"

	authRoute "github.com/jailtonjunior94/financial/internal/infrastructure/auth/web"
	categoryRoute "github.com/jailtonjunior94/financial/internal/infrastructure/category/web"
	userRoute "github.com/jailtonjunior94/financial/internal/infrastructure/user/web"
	"github.com/jailtonjunior94/financial/pkg/bundle"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/riandyrn/otelchi"
)

type ApiServe struct {
}

func NewApiServe() *ApiServe {
	return &ApiServe{}
}

func (s *ApiServe) ApiServer() {
	container := bundle.NewContainer()

	router := chi.NewRouter()
	router.Use(
		middleware.Logger,
		middleware.Heartbeat("/health"),
		middleware.Recoverer,
		middleware.SetHeader("Content-Type", "application/json"),
		otelchi.Middleware(container.Config.ServiceName, otelchi.WithChiRoutes(router)),
	)
	router.Get("/api", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authHandler := authRoute.NewAuthHandler(container.AuthUseCase)
	authRoute.NewAuthRoute(router, authRoute.WithTokenHandler(authHandler.Token))

	userHandler := userRoute.NewUserHandler(container.UserUseCase)
	userRoute.NewUserRoutes(router, userRoute.WithCreateUserHandler(userHandler.Create))

	categoryHandler := categoryRoute.NewCategoryHandler(container.CreateCategoryUseCase)
	categoryRoute.NewCategoryRoutes(router, container.MiddlewareAuth, categoryRoute.WithCreateCategoryHandler(categoryHandler.Create))

	server := http.Server{
		ReadTimeout:       time.Duration(10) * time.Second,
		ReadHeaderTimeout: time.Duration(10) * time.Second,
		Handler:           router,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", container.Config.HttpServerPort))
	if err != nil {
		panic(err)
	}
	server.Serve(listener)
}
