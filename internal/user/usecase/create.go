package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/user/domain/factories"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/encrypt"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
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
		span.AddAttributes(ctx, o11y.Error, err.Error(), o11y.Attributes{Key: "error", Value: err})
		return nil, err
	}

	hash, err := u.hash.GenerateHash(input.Password)
	if err != nil {
		span.AddAttributes(
			ctx, o11y.Error, "error generating hash",
			o11y.Attributes{Key: "e-mail", Value: input.Email},
			o11y.Attributes{Key: "error", Value: err},
		)
		return nil, err
	}

	user.SetPassword(hash)
	userCreated, err := u.repository.Insert(ctx, user)
	if err != nil {
		span.AddAttributes(
			ctx, o11y.Error, "error created user in database",
			o11y.Attributes{Key: "e-mail", Value: input.Email},
			o11y.Attributes{Key: "error", Value: err},
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
