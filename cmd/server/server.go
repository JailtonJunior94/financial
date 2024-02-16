package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/jailtonjunior94/financial/internal/category"
	"github.com/jailtonjunior94/financial/internal/user"
	"github.com/jailtonjunior94/financial/pkg/bundle"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ApiServe struct {
}

func NewApiServe() *ApiServe {
	return &ApiServe{}
}

func (s *ApiServe) ApiServer() {
	ctx := context.Background()
	ioc := bundle.NewContainer(ctx)

	/* Observability */
	tracerProvider := ioc.Observability.TracerProvider()
	defer func() {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	meterProvider := ioc.Observability.MeterProvider()
	defer func() {
		if err := meterProvider.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	router := chi.NewRouter()
	router.Use(
		middleware.Logger,
		middleware.Recoverer,
		middleware.Heartbeat("/health"),
		middleware.SetHeader("Content-Type", "application/json"),
	)

	/* Auth */
	user.RegisterAuthModule(ioc, router)
	/* User */
	user.RegisterUserModule(ioc, router)
	/* Category */
	category.RegisterCategoryModule(ioc, router)

	server := http.Server{
		ReadTimeout:       time.Duration(10) * time.Second,
		ReadHeaderTimeout: time.Duration(10) * time.Second,
		Handler:           router,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", ioc.Config.HttpServerPort))
	if err != nil {
		log.Fatal(err)
	}
	server.Serve(listener)
}
