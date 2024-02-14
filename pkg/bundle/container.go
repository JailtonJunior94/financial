package bundle

import (
	"context"
	"database/sql"
	"log"

	"github.com/jailtonjunior94/financial/configs"
	categoryInterfaces "github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	categoryRepository "github.com/jailtonjunior94/financial/internal/category/infrastructure/repository"
	category "github.com/jailtonjunior94/financial/internal/category/usecase"
	"github.com/jailtonjunior94/financial/internal/shared/web/middlewares"
	userInterfaces "github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	userRepository "github.com/jailtonjunior94/financial/internal/user/infrastructure/repository"
	user "github.com/jailtonjunior94/financial/internal/user/usecase"
	"github.com/jailtonjunior94/financial/pkg/authentication"
	"github.com/jailtonjunior94/financial/pkg/database/mysql"
	"github.com/jailtonjunior94/financial/pkg/encrypt"
	"github.com/jailtonjunior94/financial/pkg/logger"
	"github.com/jailtonjunior94/financial/pkg/observability"
	"github.com/jailtonjunior94/financial/pkg/tracing"
)

type container struct {
	DB                    *sql.DB
	Logger                logger.Logger
	Config                *configs.Config
	AuthUseCase           user.TokenUseCase
	Hash                  encrypt.HashAdapter
	CreateUserUseCase     user.CreateUserUseCase
	Jwt                   authentication.JwtAdapter
	UserRepository        userInterfaces.UserRepository
	CategoryRepository    categoryInterfaces.CategoryRepository
	MiddlewareAuth        middlewares.Authorization
	MiddlewareTracing     middlewares.TracingMiddleware
	CreateCategoryUseCase category.CreateCategoryUseCase
}

func NewContainer(ctx context.Context) *container {
	/* General Dependencies */
	config, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	dbConnection, err := mysql.NewMySqlDatabase(config)
	if err != nil {
		panic(err)
	}

	observability := observability.NewObservability(
		observability.WithServiceName(config.ServiceName),
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

	logger := logger.NewLogger()
	hash := encrypt.NewHashAdapter()
	jwt := authentication.NewJwtAdapter(logger, config)
	otelTelemetry := tracing.NewProvider(config.ServiceName, "1.0.0", config.OtelExporterEndpoint)
	middlewareTracing := middlewares.NewTracingMiddleware(otelTelemetry.GetTracer())

	/* User */
	userRepository := userRepository.NewUserRepository(dbConnection)
	userUseCase := user.NewCreateUserUseCase(logger, hash, userRepository)

	/* Auth */
	middlewareAuth := middlewares.NewAuthorization(config, jwt)
	authUseCase := user.NewTokenUseCase(tracer, config, logger, hash, jwt, userRepository)

	/* Category */
	categoryRepository := categoryRepository.NewCategoryRepository(dbConnection)
	createUserUseCase := category.NewCreateUserUseCase(logger, categoryRepository)

	return &container{
		Jwt:                   jwt,
		Config:                config,
		Hash:                  hash,
		Logger:                logger,
		DB:                    dbConnection,
		CreateUserUseCase:     userUseCase,
		AuthUseCase:           authUseCase,
		UserRepository:        userRepository,
		MiddlewareAuth:        middlewareAuth,
		MiddlewareTracing:     middlewareTracing,
		CategoryRepository:    categoryRepository,
		CreateCategoryUseCase: createUserUseCase,
	}
}
