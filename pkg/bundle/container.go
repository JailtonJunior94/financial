package bundle

import (
	"database/sql"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/domain/user/interfaces"
	"github.com/jailtonjunior94/financial/internal/infrastructure/user/repository"
	usecase "github.com/jailtonjunior94/financial/internal/usecase/user"
	mysql "github.com/jailtonjunior94/financial/pkg/database/mysql"
)

type container struct {
	Config         *configs.Config
	DB             *sql.DB
	UserRepository interfaces.UserRepository
	UserUseCase    usecase.CreateUserUseCase
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

	userRepository := repository.NewUserRepository(dbConnection)
	userUseCase := usecase.NewCreateUserUseCase(userRepository)

	return &container{
		Config:         config,
		DB:             dbConnection,
		UserRepository: userRepository,
		UserUseCase:    userUseCase,
	}
}
