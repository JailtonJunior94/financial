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
	"github.com/JailtonJunior94/devkit-go/pkg/logger"
	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
)

type Container struct {
	DB                     *sql.DB
	Logger                 logger.Logger
	Config                 *configs.Config
	Jwt                    auth.JwtAdapter
	Hash                   encrypt.HashAdapter
	Observability          o11y.Observability
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

	observability := o11y.NewObservability(
		o11y.WithServiceName(config.O11yConfig.ServiceName),
		o11y.WithServiceVersion(config.O11yConfig.ServiceVersion),
		o11y.WithResource(),
		o11y.WithTracerProvider(ctx, config.O11yConfig.ExporterEndpoint),
		o11y.WithMeterProvider(ctx, config.O11yConfig.ExporterEndpoint),
		o11y.WithLoggerProvider(ctx, config.O11yConfig.ExporterEndpointHTTP),
	)

	logger := logger.NewLogger()
	hash := encrypt.NewHashAdapter()
	jwt := auth.NewJwtAdapter(config, observability)
	middlewareAuth := middlewares.NewAuthorization(config, jwt)
	panicRecoverMiddleware := middlewares.NewPanicRecoverMiddleware(observability)

	return &Container{
		DB:                     db,
		Jwt:                    jwt,
		Hash:                   hash,
		Logger:                 logger,
		Config:                 config,
		MiddlewareAuth:         middlewareAuth,
		Observability:          observability,
		PanicRecoverMiddleware: panicRecoverMiddleware,
	}
}
