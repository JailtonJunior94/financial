package bundle

import (
	"context"
	"database/sql"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/database/mysql"

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
		panic(err)
	}

	db, err := mysql.NewMySqlDatabase(config)
	if err != nil {
		panic(err)
	}

	observability := o11y.NewObservability(
		o11y.WithServiceName(config.ServiceName),
		o11y.WithServiceVersion(config.ServiceVersion),
		o11y.WithResource(),
		o11y.WithLoggerProvider(ctx, config.OtelExporterEndpoint),
		o11y.WithTracerProvider(ctx, config.OtelExporterEndpoint),
		o11y.WithMeterProvider(ctx, config.OtelExporterEndpoint),
	)

	logger := logger.NewLogger()
	hash := encrypt.NewHashAdapter()
	jwt := auth.NewJwtAdapter(config, observability)
	middlewareAuth := middlewares.NewAuthorization(config, jwt)
	panicRecoverMiddleware := middlewares.NewPanicRecoverMiddleware(observability)

	return &Container{
		DB:                     db,
		Logger:                 logger,
		Config:                 config,
		Jwt:                    jwt,
		Hash:                   hash,
		MiddlewareAuth:         middlewareAuth,
		Observability:          observability,
		PanicRecoverMiddleware: panicRecoverMiddleware,
	}
}
