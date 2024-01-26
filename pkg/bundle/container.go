package bundle

import (
	"database/sql"

	"github.com/jailtonjunior94/financial/configs"
	categoryInterface "github.com/jailtonjunior94/financial/internal/domain/category/interfaces"
	userInterface "github.com/jailtonjunior94/financial/internal/domain/user/interfaces"
	cr "github.com/jailtonjunior94/financial/internal/infrastructure/category/repository"
	ur "github.com/jailtonjunior94/financial/internal/infrastructure/user/repository"
	"github.com/jailtonjunior94/financial/internal/infrastructure/web/middlewares"
	"github.com/jailtonjunior94/financial/internal/usecase/auth"
	"github.com/jailtonjunior94/financial/internal/usecase/category"
	"github.com/jailtonjunior94/financial/internal/usecase/user"
	"github.com/jailtonjunior94/financial/pkg/authentication"
	mysql "github.com/jailtonjunior94/financial/pkg/database/mysql"
	"github.com/jailtonjunior94/financial/pkg/encrypt"
	"github.com/jailtonjunior94/financial/pkg/logger"
	"github.com/jailtonjunior94/financial/pkg/tracing"
)

type container struct {
	DB                    *sql.DB
	Logger                logger.Logger
	Config                *configs.Config
	AuthUseCase           auth.TokenUseCase
	Hash                  encrypt.HashAdapter
	UserUseCase           user.CreateUserUseCase
	Jwt                   authentication.JwtAdapter
	UserRepository        userInterface.UserRepository
	CategoryRepository    categoryInterface.CategoryRepository
	MiddlewareAuth        middlewares.Authorization
	MiddlewareTracing     middlewares.TracingMiddleware
	CreateCategoryUseCase category.CreateCategoryUseCase
}

func NewContainer() *container {
	/* General Dependencies */
	config, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	dbConnection, err := mysql.NewMySqlDatabase(config)
	if err != nil {
		panic(err)
	}

	logger := logger.NewLogger()
	hash := encrypt.NewHashAdapter()
	jwt := authentication.NewJwtAdapter(logger, config)
	otelTelemetry := tracing.NewProvider(config.ServiceName, "1.0.0", config.OtelExporterEndpoint)
	middlewareTracing := middlewares.NewTracingMiddleware(otelTelemetry.GetTracer())

	/* User */
	userRepository := ur.NewUserRepository(dbConnection)
	userUseCase := user.NewCreateUserUseCase(logger, hash, userRepository)

	/* Auth */
	middlewareAuth := middlewares.NewAuthorization(config, jwt)
	authUseCase := auth.NewTokenUseCase(config, logger, hash, jwt, userRepository)

	/* Category */
	categoryRepository := cr.NewCategoryRepository(dbConnection)
	createUserUseCase := category.NewCreateUserUseCase(logger, categoryRepository)

	return &container{
		Jwt:                   jwt,
		Config:                config,
		Hash:                  hash,
		Logger:                logger,
		DB:                    dbConnection,
		UserUseCase:           userUseCase,
		AuthUseCase:           authUseCase,
		UserRepository:        userRepository,
		MiddlewareAuth:        middlewareAuth,
		MiddlewareTracing:     middlewareTracing,
		CategoryRepository:    categoryRepository,
		CreateCategoryUseCase: createUserUseCase,
	}
}
