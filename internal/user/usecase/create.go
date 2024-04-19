package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/user/domain/factories"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/encrypt"
	"github.com/jailtonjunior94/financial/pkg/logger"
)

type (
	CreateUserUseCase interface {
		Execute(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error)
	}

	createUserUseCase struct {
		logger     logger.Logger
		hash       encrypt.HashAdapter
		repository interfaces.UserRepository
	}
)

func NewCreateUserUseCase(
	logger logger.Logger,
	hash encrypt.HashAdapter,
	repository interfaces.UserRepository,
) CreateUserUseCase {
	return &createUserUseCase{
		logger:     logger,
		hash:       hash,
		repository: repository,
	}
}

func (u *createUserUseCase) Execute(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
	user, err := factories.CreateUser(input.Name, input.Email)
	if err != nil {
		u.logger.Warn("error parsing user", logger.Field{Key: "warning", Value: err})
		return nil, err
	}

	hash, err := u.hash.GenerateHash(input.Password)
	if err != nil {
		u.logger.Error("error generating hash",
			logger.Field{Key: "e-mail", Value: input.Email},
			logger.Field{Key: "error", Value: err},
		)
		return nil, err
	}

	user.SetPassword(hash)
	userCreated, err := u.repository.Insert(ctx, user)
	if err != nil {
		u.logger.Error("error created user in database",
			logger.Field{Key: "e-mail", Value: input.Email},
			logger.Field{Key: "error", Value: err},
		)
		return nil, err
	}

	return &CreateUserOutput{
		ID:        userCreated.ID.String(),
		Name:      userCreated.Name.String(),
		Email:     userCreated.Email.String(),
		CreatedAt: userCreated.CreatedAt,
	}, nil
}
