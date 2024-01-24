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
)

type container struct {
	Config         *configs.Config
	DB             *sql.DB
	Hash           encrypt.HashAdapter
	Jwt            authentication.JwtAdapter
	UserRepository interfaces.UserRepository
	UserUseCase    user.CreateUserUseCase
	AuthUseCase    auth.TokenUseCase
	MiddlewareAuth middlewares.Authorization
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

	hash := encrypt.NewHashAdapter()
	jwt := authentication.NewJwtAdapter(config)
	middlewareAuth := middlewares.NewAuthorization(config)
	userRepository := repository.NewUserRepository(dbConnection)
	userUseCase := user.NewCreateUserUseCase(hash, userRepository)
	authUseCase := auth.NewTokenUseCase(hash, jwt, userRepository)

	return &container{
		Config:         config,
		Hash:           hash,
		Jwt:            jwt,
		DB:             dbConnection,
		UserRepository: userRepository,
		UserUseCase:    userUseCase,
		AuthUseCase:    authUseCase,
		MiddlewareAuth: middlewareAuth,
	}
}
