package user

import (
	"github.com/jailtonjunior94/financial/internal/domain/user/entity"
	"github.com/jailtonjunior94/financial/internal/domain/user/interfaces"
	"github.com/jailtonjunior94/financial/pkg/encrypt"
	"github.com/jailtonjunior94/financial/pkg/logger"
)

type CreateUserUseCase interface {
	Execute(input *CreateUserInput) (*CreateUserOutput, error)
}

type createUserUseCase struct {
	logger     logger.Logger
	hash       encrypt.HashAdapter
	repository interfaces.UserRepository
}

func NewCreateUserUseCase(logger logger.Logger, hash encrypt.HashAdapter, repository interfaces.UserRepository) CreateUserUseCase {
	return &createUserUseCase{logger: logger, hash: hash, repository: repository}
}

func (u *createUserUseCase) Execute(input *CreateUserInput) (*CreateUserOutput, error) {
	newUser, err := entity.NewUser(input.Name, input.Email)
	if err != nil {
		return nil, err
	}

	u.logger.Info("generating hash", logger.Field{Key: "password", Value: input.Password})

	hash, err := u.hash.GenerateHash(input.Password)
	if err != nil {
		return nil, err
	}

	newUser.SetPassword(hash)
	user, err := u.repository.Create(newUser)
	if err != nil {
		u.logger.Error("error creating user", logger.Field{Key: "error", Value: err.Error()})
		return nil, err
	}

	return &CreateUserOutput{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}
