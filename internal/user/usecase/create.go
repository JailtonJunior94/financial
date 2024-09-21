package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/user/domain/factories"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/encrypt"
	"github.com/jailtonjunior94/financial/pkg/o11y"
)

type (
	CreateUserUseCase interface {
		Execute(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error)
	}

	createUserUseCase struct {
		o11y       o11y.Observability
		hash       encrypt.HashAdapter
		repository interfaces.UserRepository
	}
)

func NewCreateUserUseCase(
	o11y o11y.Observability,
	hash encrypt.HashAdapter,
	repository interfaces.UserRepository,
) CreateUserUseCase {
	return &createUserUseCase{
		o11y:       o11y,
		hash:       hash,
		repository: repository,
	}
}

func (u *createUserUseCase) Execute(ctx context.Context, input *CreateUserInput) (*CreateUserOutput, error) {
	ctx, span := u.o11y.Start(ctx, "create_user_usecase.execute")
	defer span.End()

	user, err := factories.CreateUser(input.Name, input.Email)
	if err != nil {
		span.AddStatus(o11y.Error, err.Error())
		span.AddAttributes(
			o11y.Attributes{Key: "error", Value: err.Error()},
		)
		return nil, err
	}

	hash, err := u.hash.GenerateHash(input.Password)
	if err != nil {
		span.AddStatus(o11y.Error, "error generating hash")
		span.AddAttributes(
			o11y.Attributes{Key: "e-mail", Value: input.Email},
			o11y.Attributes{Key: "error", Value: err.Error()},
		)
		return nil, err
	}

	user.SetPassword(hash)
	userCreated, err := u.repository.Insert(ctx, user)
	if err != nil {
		span.AddStatus(o11y.Error, "error created user in database")
		span.AddAttributes(
			o11y.Attributes{Key: "e-mail", Value: input.Email},
			o11y.Attributes{Key: "error", Value: err.Error()},
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
