package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jailtonjunior94/financial/internal/category"
	"github.com/jailtonjunior94/financial/internal/user"
	"github.com/jailtonjunior94/financial/pkg/bundle"
	"github.com/jailtonjunior94/financial/pkg/responses"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type apiServer struct {
}

func NewApiServer() *apiServer {
	return &apiServer{}
}

func (s *apiServer) Server() {
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

	/* Close DBConnection */
	defer func() {
		if err := ioc.DB.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	router := chi.NewRouter()
	router.Use(
		middleware.RealIP,
		middleware.RequestID,
		middleware.Recoverer,
		ioc.MetricMiddleware.Metrics,
		middleware.AllowContentType("application/json", "application/x-www-form-urlencoded"),
		middleware.SetHeader("Content-Type", "application/json"),
	)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := ioc.DB.Ping(); err != nil {
			responses.Error(w, http.StatusInternalServerError, "database error connection failed or database is not running")
			return
		}
		responses.JSON(w, http.StatusOK, map[string]interface{}{"status": "ok"})
	})

	/* Auth */
	user.RegisterAuthModule(ioc, router)
	/* User */
	user.RegisterUserModule(ioc, router)
	/* Category */
	category.RegisterCategoryModule(ioc, router)

	/* Graceful shutdown */
	server := http.Server{
		ReadTimeout:       time.Duration(10) * time.Second,
		ReadHeaderTimeout: time.Duration(10) * time.Second,
		Handler:           router,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", ioc.Config.HttpServerPort))
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	s.gracefulShutdown(&server)
}

func (s *apiServer) gracefulShutdown(server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatal(err)
	}
}
