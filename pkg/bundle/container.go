package bundle

import (
	"database/sql"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/domain/user/interfaces"
	"github.com/jailtonjunior94/financial/internal/infrastructure/user/repository"
	"github.com/jailtonjunior94/financial/internal/infrastructure/web/middlewares"
	"github.com/jailtonjunior94/financial/internal/usecase/auth"
	"github.com/jailtonjunior94/financial/internal/usecase/user"
	"github.com/jailtonjunior94/financial/pkg/authentication"
	mysql "github.com/jailtonjunior94/financial/pkg/database/mysql"
	"github.com/jailtonjunior94/financial/pkg/encrypt"
	"github.com/jailtonjunior94/financial/pkg/logger"
	"github.com/jailtonjunior94/financial/pkg/tracing"
)

type container struct {
	DB                *sql.DB
	Logger            logger.Logger
	Config            *configs.Config
	AuthUseCase       auth.TokenUseCase
	Hash              encrypt.HashAdapter
	UserUseCase       user.CreateUserUseCase
	Jwt               authentication.JwtAdapter
	UserRepository    interfaces.UserRepository
	MiddlewareAuth    middlewares.Authorization
	MiddlewareTracing middlewares.TracingMiddleware
}

func NewContainer() *container {
	config, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	dbConnection, err := mysql.NewMySqlDatabase(config)
	if err != nil {
		panic(err)
	}

	otelTelemetry := tracing.NewProvider(config.ServiceName, "1.0.0", config.OtelExporterEndpoint)
	logger := logger.NewLogger()
	hash := encrypt.NewHashAdapter()
	jwt := authentication.NewJwtAdapter(config)
	middlewareAuth := middlewares.NewAuthorization(config)
	userRepository := repository.NewUserRepository(dbConnection)
	middlewareTracing := middlewares.NewTracingMiddleware(otelTelemetry.GetTracer())
	authUseCase := auth.NewTokenUseCase(hash, jwt, userRepository)
	userUseCase := user.NewCreateUserUseCase(logger, hash, userRepository)

	return &container{
		Jwt:               jwt,
		Config:            config,
		Hash:              hash,
		Logger:            logger,
		DB:                dbConnection,
		UserUseCase:       userUseCase,
		AuthUseCase:       authUseCase,
		UserRepository:    userRepository,
		MiddlewareAuth:    middlewareAuth,
		MiddlewareTracing: middlewareTracing,
	}
}
