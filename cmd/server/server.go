package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jailtonjunior94/financial/internal/category"
	"github.com/jailtonjunior94/financial/internal/user"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/bundle"

	"github.com/JailtonJunior94/devkit-go/pkg/httpserver"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
)

func Run() {
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

	healthRoute := httpserver.NewRoute(http.MethodGet, "/health", func(w http.ResponseWriter, r *http.Request) error {
		if err := ioc.DB.Ping(); err != nil {
			return err
		}
		responses.JSON(w, http.StatusOK, map[string]interface{}{"status": "ok"})
		return nil
	})

	routes := []httpserver.Route{healthRoute}
	authRoutes := user.RegisterAuthModule(ioc)
	userRoutes := user.RegisterUserModule(ioc)
	categoryRoutes := category.RegisterCategoryModule(ioc)

	routes = append(routes, authRoutes...)
	routes = append(routes, userRoutes...)
	routes = append(routes, categoryRoutes...)

	server := httpserver.New(
		httpserver.WithPort(ioc.Config.HttpServerPort),
		httpserver.WithRoutes(routes...),
		httpserver.WithErrorHandler(middlewares.ErrorHandler),
		httpserver.WithMiddlewares(
			httpserver.RequestID,
		),
	)

	shutdown := server.Run()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := <-server.ShutdownListener(); err != nil && err != http.ErrServerClosed {
			interrupt <- syscall.SIGTERM
		}
	}()

	<-interrupt
	if err := shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}
}
