package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/auth"
	financialErrors "github.com/jailtonjunior94/financial/pkg/error"

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
	o11y       o11y.Observability
}

func NewTokenUseCase(
	config *configs.Config,
	o11y o11y.Observability,
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
	ctx, span := u.o11y.Start(ctx, "create_user_usecase.execute")
	defer span.End()

	user, err := u.repository.FindByEmail(ctx, input.Email)
	if err != nil {
		span.AddAttributes(
			ctx, o11y.Error, "error find user by e-mail",
			o11y.Attributes{Key: EmailKey, Value: input.Email},
			o11y.Attributes{Key: "error", Value: err},
		)
		return nil, err
	}

	if user == nil {
		span.AddAttributes(
			ctx, o11y.Error, "user not found",
			o11y.Attributes{Key: EmailKey, Value: input.Email},
			o11y.Attributes{Key: "error", Value: err},
		)
		return nil, financialErrors.ErrUserNotFound
	}

	if !u.hash.CheckHash(user.Password, input.Password) {
		span.AddAttributes(
			ctx, o11y.Error, "error checking hash",
			o11y.Attributes{Key: EmailKey, Value: input.Email},
			o11y.Attributes{Key: "error", Value: err},
		)
		return nil, financialErrors.ErrCheckHash
	}

	token, err := u.jwt.GenerateToken(ctx, user.ID.String(), user.Email.String())
	if err != nil {
		span.AddAttributes(
			ctx, o11y.Error, "error generate token",
			o11y.Attributes{Key: EmailKey, Value: input.Email},
			o11y.Attributes{Key: "error", Value: err},
		)
		return nil, err
	}
	return dtos.NewAuthOutput(token, u.config.AuthConfig.AuthTokenDuration), nil
}
