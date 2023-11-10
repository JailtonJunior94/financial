package bundle

import (
	"database/sql"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/domain/user/interfaces"
	"github.com/jailtonjunior94/financial/internal/infrastructure/user/repository"
	usecase "github.com/jailtonjunior94/financial/internal/usecase/user"
	mysql "github.com/jailtonjunior94/financial/pkg/database/mysql"
	"github.com/jailtonjunior94/financial/pkg/encrypt"
)

type container struct {
	Config         *configs.Config
	DB             *sql.DB
	UserRepository interfaces.UserRepository
	UserUseCase    usecase.CreateUserUseCase
	Hash           encrypt.HashAdapter
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
	userRepository := repository.NewUserRepository(dbConnection)
	userUseCase := usecase.NewCreateUserUseCase(hash, userRepository)

	return &container{
		Config:         config,
		DB:             dbConnection,
		Hash:           hash,
		UserRepository: userRepository,
		UserUseCase:    userUseCase,
	}
}
