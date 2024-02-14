package server

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	authRoute "github.com/jailtonjunior94/financial/internal/infrastructure/auth/web"
	categoryRoute "github.com/jailtonjunior94/financial/internal/infrastructure/category/web"
	userRoute "github.com/jailtonjunior94/financial/internal/infrastructure/user/web"
	"github.com/jailtonjunior94/financial/pkg/bundle"
	"github.com/jailtonjunior94/financial/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ApiServe struct {
}

func NewApiServe() *ApiServe {
	return &ApiServe{}
}

func (s *ApiServe) ApiServer() {
	container := bundle.NewContainer()

	ctx := context.Background()
	observability := observability.NewObservability(
		observability.WithServiceName(container.Config.ServiceName),
		observability.WithServiceVersion("1.0.0"),
		observability.WithResource(),
		observability.WithTracerProvider(ctx, "localhost:4317"),
		observability.WithMeterProvider(ctx, "localhost:4317"),
	)

	tracerProvider := observability.TracerProvider()
	defer func() {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	meterProvider := observability.MeterProvider()
	defer func() {
		if err := meterProvider.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	tracer := observability.Tracer()

	router := chi.NewRouter()
	router.Use(
		middleware.Logger,
		middleware.Recoverer,
		middleware.Heartbeat("/health"),
		middleware.SetHeader("Content-Type", "application/json"),
	)

	router.Get("/api", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "roll")
		defer span.End()

		meter := meterProvider.Meter("sample")
		rollCnt, _ := meter.Int64Counter("dice.rolls",
			metric.WithDescription("The number of rolls by roll value"),
			metric.WithUnit("{roll}"),
		)

		roll := 1 + rand.Intn(6)
		rollValueAttr := attribute.Int("roll.value", roll)
		rollCnt.Add(ctx, 1, metric.WithAttributes(rollValueAttr))

		w.WriteHeader(http.StatusOK)
		fmt.Println(ctx)
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
