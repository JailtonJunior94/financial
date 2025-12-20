package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/auth"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/encrypt"
	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
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
	o11y       o11y.Telemetry
}

func NewTokenUseCase(
	config *configs.Config,
	o11y o11y.Telemetry,
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
	ctx, span := u.o11y.Tracer().Start(ctx, "create_user_usecase.execute")
	defer span.End()

	user, err := u.repository.FindByEmail(ctx, input.Email)
	if err != nil {
		span.AddEvent(
			"error finding user by e-mail",
			o11y.Attribute{Key: "e-mail", Value: input.Email},
			o11y.Attribute{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, err, "error finding user by e-mail", o11y.Field{Key: "e-mail", Value: input.Email})
		return nil, err
	}

	if user == nil {
		span.AddEvent(
			"user not found",
			o11y.Attribute{Key: EmailKey, Value: input.Email},
			o11y.Attribute{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, err, "user not found", o11y.Field{Key: EmailKey, Value: input.Email})
		return nil, customErrors.New("user or password invalid", fmt.Errorf("token_usecase: %v", err))
	}

	if !u.hash.CheckHash(user.Password, input.Password) {
		span.AddEvent(
			"error checking hash",
			o11y.Attribute{Key: EmailKey, Value: input.Email},
			o11y.Attribute{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, err, "error checking hash", o11y.Field{Key: EmailKey, Value: input.Email})
		return nil, customErrors.New("user or password invalid", fmt.Errorf("token_usecase: %v", err))
	}

	token, err := u.jwt.GenerateToken(ctx, user.ID.String(), user.Email.String())
	if err != nil {
		span.AddEvent(
			"error generate token",
			o11y.Attribute{Key: EmailKey, Value: input.Email},
			o11y.Attribute{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, err, "error generate token", o11y.Field{Key: EmailKey, Value: input.Email})
		return nil, err
	}
	return dtos.NewAuthOutput(token, u.config.AuthConfig.AuthTokenDuration), nil
}
