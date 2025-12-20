package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/domain/factories"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/encrypt"
	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
)

type (
	CreateUserUseCase interface {
		Execute(ctx context.Context, input *dtos.CreateUserInput) (*dtos.CreateUserOutput, error)
	}

	createUserUseCase struct {
		o11y       o11y.Telemetry
		hash       encrypt.HashAdapter
		repository interfaces.UserRepository
	}
)

func NewCreateUserUseCase(
	o11y o11y.Telemetry,
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

	user, err := factories.CreateUser(input.Name, input.Email)
	if err != nil {
		span.AddEvent(
			"error creating user entity",
			o11y.Attribute{Key: "e-mail", Value: input.Email},
			o11y.Attribute{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, err, "error creating user entity", o11y.Field{Key: "e-mail", Value: input.Email})
		return nil, customErrors.New("error creating user", fmt.Errorf("create_user_usecase: %v", err))
	}

	hash, err := u.hash.GenerateHash(input.Password)
	if err != nil {
		span.AddEvent(
			"error generating hash",
			o11y.Attribute{Key: "e-mail", Value: input.Email},
			o11y.Attribute{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, err, "error generating hash", o11y.Field{Key: "e-mail", Value: input.Email})
		return nil, err
	}

	if err := user.SetPassword(hash); err != nil {
		span.AddEvent("error setting password", o11y.Attribute{Key: "error", Value: err})
		u.o11y.Logger().Error(ctx, err, "error setting password")
		return nil, err
	}

	userCreated, err := u.repository.Insert(ctx, user)
	if err != nil {
		span.AddEvent(
			"error inserting user into repository",
			o11y.Attribute{Key: "e-mail", Value: input.Email},
			o11y.Attribute{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, err, "error inserting user into repository", o11y.Field{Key: "e-mail", Value: input.Email})
		return nil, err
	}

	return &dtos.CreateUserOutput{
		ID:        userCreated.ID.String(),
		Name:      userCreated.Name.String(),
		Email:     userCreated.Email.String(),
		CreatedAt: userCreated.CreatedAt,
	}, nil
}
