package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	categoryRoute "github.com/jailtonjunior94/financial/internal/category/infrastructure/web"
	userRoute "github.com/jailtonjunior94/financial/internal/user/infrastructure/web"
	"github.com/jailtonjunior94/financial/pkg/bundle"
	"github.com/riandyrn/otelchi"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ApiServe struct {
}

func NewApiServe() *ApiServe {
	return &ApiServe{}
}

func (s *ApiServe) ApiServer() {
	ioc := bundle.NewContainer(context.Background())

	router := chi.NewRouter()
	router.Use(
		middleware.Logger,
		middleware.Recoverer,
		middleware.Heartbeat("/health"),
		middleware.SetHeader("Content-Type", "application/json"),
		otelchi.Middleware(ioc.Config.ServiceName, otelchi.WithChiRoutes(router)),
	)

	authHandler := userRoute.NewAuthHandler(ioc.AuthUseCase)
	userRoute.NewAuthRoute(router, userRoute.WithTokenHandler(authHandler.Token))

	userHandler := userRoute.NewUserHandler(ioc.CreateUserUseCase)
	userRoute.NewUserRoutes(router, userRoute.WithCreateUserHandler(userHandler.Create))

	categoryHandler := categoryRoute.NewCategoryHandler(ioc.CreateCategoryUseCase)
	categoryRoute.NewCategoryRoutes(router, ioc.MiddlewareAuth, categoryRoute.WithCreateCategoryHandler(categoryHandler.Create))

	server := http.Server{
		ReadTimeout:       time.Duration(10) * time.Second,
		ReadHeaderTimeout: time.Duration(10) * time.Second,
		Handler:           router,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", ioc.Config.HttpServerPort))
	if err != nil {
		panic(err)
	}
	server.Serve(listener)
}
