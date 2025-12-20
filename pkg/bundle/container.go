package bundle

import (
	"context"
	"database/sql"
	"log"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/database/postgres"

	"github.com/JailtonJunior94/devkit-go/pkg/encrypt"
	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
)

type Container struct {
	DB                     *sql.DB
	Config                 *configs.Config
	Jwt                    auth.JwtAdapter
	Hash                   encrypt.HashAdapter
	Telemetry              o11y.Telemetry
	MiddlewareAuth         middlewares.Authorization
	PanicRecoverMiddleware middlewares.PanicRecoverMiddleware
}

func NewContainer(ctx context.Context) *Container {
	config, err := configs.LoadConfig(".")
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	db, err := postgres.NewPostgresDatabase(config)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	resource, err := o11y.NewServiceResource(ctx, config.O11yConfig.ServiceName, config.O11yConfig.ServiceVersion, "development")
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	tracer, tracerShutdown, err := o11y.NewTracerWithOptions(ctx,
		o11y.WithTracerEndpoint(config.O11yConfig.ExporterEndpoint),
		o11y.WithTracerServiceName(config.O11yConfig.ServiceName),
		o11y.WithTracerResource(resource),
		o11y.WithTracerInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to create tracer: %v", err)
	}

	metrics, metricsShutdown, err := o11y.NewMetricsWithOptions(ctx,
		o11y.WithMetricsEndpoint(config.O11yConfig.ExporterEndpoint),
		o11y.WithMetricsServiceName(config.O11yConfig.ServiceName),
		o11y.WithMetricsResource(resource),
		o11y.WithMetricsInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to create metrics: %v", err)
	}

	logger, loggerShutdown, err := o11y.NewLoggerWithOptions(ctx,
		o11y.WithLoggerEndpoint(config.O11yConfig.ExporterEndpointHTTP),
		o11y.WithLoggerServiceName(config.O11yConfig.ServiceName),
		o11y.WithLoggerResource(resource),
		o11y.WithLoggerInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	telemetry, err := o11y.NewTelemetry(tracer, metrics, logger, tracerShutdown, metricsShutdown, loggerShutdown)
	if err != nil {
		log.Fatalf("failed to create telemetry: %v", err)
	}

	hash := encrypt.NewHashAdapter()
	jwt := auth.NewJwtAdapter(config, telemetry)
	middlewareAuth := middlewares.NewAuthorization(config, jwt)
	panicRecoverMiddleware := middlewares.NewPanicRecoverMiddleware(telemetry)

	return &Container{
		DB:                     db,
		Jwt:                    jwt,
		Hash:                   hash,
		Config:                 config,
		MiddlewareAuth:         middlewareAuth,
		Telemetry:              telemetry,
		PanicRecoverMiddleware: panicRecoverMiddleware,
	}
}
