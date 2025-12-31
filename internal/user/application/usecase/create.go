package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/domain/factories"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/encrypt"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type (
	CreateUserUseCase interface {
		Execute(ctx context.Context, input *dtos.CreateUserInput) (*dtos.CreateUserOutput, error)
	}

	createUserUseCase struct {
		o11y       observability.Observability
		hash       encrypt.HashAdapter
		repository interfaces.UserRepository
	}
)

func NewCreateUserUseCase(
	o11y observability.Observability,
	hash encrypt.HashAdapter,
	repository interfaces.UserRepository,
) CreateUserUseCase {
	return &createUserUseCase{
		o11y:       o11y,
		hash:       hash,
		repository: repository,
	}
}

func (u *createUserUseCase) Execute(ctx context.Context, input *dtos.CreateUserInput) (*dtos.CreateUserOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "create_user_usecase.execute")
	defer span.End()

	span.AddEvent("creating user",
		observability.Field{Key: "user.email", Value: input.Email},
		observability.Field{Key: "user.name", Value: input.Name},
	)

	// Validate input
	if input.Password == "" {
		span.RecordError(customErrors.ErrPasswordIsRequired)
		return nil, customErrors.ErrPasswordIsRequired
	}

	user, err := factories.CreateUser(input.Name, input.Email)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	hash, err := u.hash.GenerateHash(input.Password)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("generating password hash: %w", err)
	}

	if err := user.SetPassword(hash); err != nil {
		span.RecordError(err)
		return nil, err
	}

	userCreated, err := u.repository.Insert(ctx, user)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &dtos.CreateUserOutput{
		ID:        userCreated.ID.String(),
		Name:      userCreated.Name.String(),
		Email:     userCreated.Email.String(),
		CreatedAt: userCreated.CreatedAt,
	}, nil
}
