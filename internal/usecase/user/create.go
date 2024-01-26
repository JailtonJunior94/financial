package user

import (
	"github.com/jailtonjunior94/financial/internal/domain/user/entity"
	"github.com/jailtonjunior94/financial/internal/domain/user/interfaces"
	"github.com/jailtonjunior94/financial/pkg/encrypt"
	"github.com/jailtonjunior94/financial/pkg/logger"
)

type (
	CreateUserUseCase interface {
		Execute(input *CreateUserInput) (*CreateUserOutput, error)
	}

	createUserUseCase struct {
		logger     logger.Logger
		hash       encrypt.HashAdapter
		repository interfaces.UserRepository
	}
)

func NewCreateUserUseCase(logger logger.Logger, hash encrypt.HashAdapter, repository interfaces.UserRepository) CreateUserUseCase {
	return &createUserUseCase{logger: logger, hash: hash, repository: repository}
}

func (u *createUserUseCase) Execute(input *CreateUserInput) (*CreateUserOutput, error) {
	newUser, err := entity.NewUser(input.Name, input.Email)
	if err != nil {
		u.logger.Warn("error parsing user", logger.Field{Key: "warning", Value: err.Error()})
		return nil, err
	}

	hash, err := u.hash.GenerateHash(input.Password)
	if err != nil {
		u.logger.Error("error generating hash",
			logger.Field{Key: "e-mail", Value: input.Email},
			logger.Field{Key: "error", Value: err.Error()},
		)
		return nil, err
	}

	newUser.SetPassword(hash)
	user, err := u.repository.Create(newUser)
	if err != nil {
		u.logger.Error("error created user in database",
			logger.Field{Key: "e-mail", Value: input.Email},
			logger.Field{Key: "error", Value: err.Error()},
		)
		return nil, err
	}

	return &CreateUserOutput{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}
