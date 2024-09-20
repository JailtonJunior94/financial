package bundle

import (
	"context"
	"database/sql"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/shared/http/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/database/postgres"
	"github.com/jailtonjunior94/financial/pkg/encrypt"
	"github.com/jailtonjunior94/financial/pkg/logger"
	"github.com/jailtonjunior94/financial/pkg/observability"
)

type Container struct {
	DB             *sql.DB
	Logger         logger.Logger
	Config         *configs.Config
	Jwt            auth.JwtAdapter
	Hash           encrypt.HashAdapter
	MiddlewareAuth middlewares.Authorization
	Observability  observability.Observability
}

func NewContainer(ctx context.Context) *Container {
	config, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	db, err := postgres.NewPostgresDatabase(config)
	if err != nil {
		panic(err)
	}

	observability := observability.NewObservability(
		observability.WithServiceName(config.ServiceName),
		observability.WithServiceVersion(config.ServiceVersion),
		observability.WithResource(),
		observability.WithLoggerProvider(ctx, config.OtelExporterEndpoint),
		observability.WithTracerProvider(ctx, config.OtelExporterEndpoint),
		observability.WithMeterProvider(ctx, config.OtelExporterEndpoint),
	)

	logger := logger.NewLogger()
	hash := encrypt.NewHashAdapter()
	jwt := auth.NewJwtAdapter(config, observability)
	middlewareAuth := middlewares.NewAuthorization(config, jwt)

	return &Container{
		DB:             db,
		Logger:         logger,
		Config:         config,
		Jwt:            jwt,
		Hash:           hash,
		MiddlewareAuth: middlewareAuth,
		Observability:  observability,
	}
}
