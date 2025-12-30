package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/auth"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/encrypt"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

const (
	EmailKey = "email"
	ErrorKey = "error"
)

type TokenUseCase interface {
	Execute(ctx context.Context, input *dtos.AuthInput) (*dtos.AuthOutput, error)
}

type tokenUseCase struct {
	jwt        auth.JwtAdapter
	config     *configs.Config
	hash       encrypt.HashAdapter
	repository interfaces.UserRepository
	o11y       observability.Observability
}

func NewTokenUseCase(
	config *configs.Config,
	o11y observability.Observability,
	hash encrypt.HashAdapter,
	jwt auth.JwtAdapter,
	repository interfaces.UserRepository,
) TokenUseCase {
	return &tokenUseCase{
		config:     config,
		hash:       hash,
		jwt:        jwt,
		repository: repository,
		o11y:       o11y,
	}
}

func (u *tokenUseCase) Execute(ctx context.Context, input *dtos.AuthInput) (*dtos.AuthOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "token_usecase.execute")
	defer span.End()

	// Validate input
	if input.Email == "" || input.Password == "" {
		validationErr := customErrors.New("email and password are required", customErrors.ErrPasswordIsRequired)
		span.AddEvent("invalid credentials", observability.Field{Key: "error", Value: validationErr})
		u.o11y.Logger().Error(ctx, "email and password are required", observability.Error(validationErr))
		return nil, validationErr
	}

	user, err := u.repository.FindByEmail(ctx, input.Email)
	if err != nil {
		span.AddEvent(
			"error finding user by e-mail",
			observability.Field{Key: "e-mail", Value: input.Email},
			observability.Field{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, "error finding user by e-mail", observability.Error(err), observability.String("e-mail", input.Email))
		return nil, err
	}

	if user == nil {
		userNotFoundErr := customErrors.ErrUserNotFound
		span.AddEvent(
			"user not found",
			observability.Field{Key: EmailKey, Value: input.Email},
			observability.Field{Key: "error", Value: userNotFoundErr},
		)
		u.o11y.Logger().Error(ctx, "user not found", observability.Error(userNotFoundErr), observability.String(EmailKey, input.Email))
		return nil, customErrors.New("user or password invalid", userNotFoundErr)
	}

	if !u.hash.CheckHash(user.Password, input.Password) {
		invalidPasswordErr := customErrors.ErrCheckHash
		span.AddEvent(
			"invalid password",
			observability.Field{Key: EmailKey, Value: input.Email},
			observability.Field{Key: "error", Value: invalidPasswordErr},
		)
		u.o11y.Logger().Error(ctx, "invalid password", observability.Error(invalidPasswordErr), observability.String(EmailKey, input.Email))
		return nil, customErrors.New("user or password invalid", invalidPasswordErr)
	}

	token, err := u.jwt.GenerateToken(ctx, user.ID.String(), user.Email.String())
	if err != nil {
		span.AddEvent(
			"error generate token",
			observability.Field{Key: EmailKey, Value: input.Email},
			observability.Field{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, "error generate token", observability.Error(err), observability.String(EmailKey, input.Email))
		return nil, err
	}
	return dtos.NewAuthOutput(token, u.config.AuthConfig.AuthTokenDuration), nil
}
