package usecase

import (
	"context"
	"errors"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/encrypt"
	"github.com/jailtonjunior94/financial/pkg/observability"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrCheckHash    = errors.New("error checking hash")
)

const (
	EmailKey = "email"
	ErrorKey = "error"
)

type TokenUseCase interface {
	Execute(ctx context.Context, input *AuthInput) (*AuthOutput, error)
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

func (u *tokenUseCase) Execute(ctx context.Context, input *AuthInput) (*AuthOutput, error) {
	ctx, span := u.o11y.Start(ctx, "create_user_usecase.execute")
	defer span.End()

	user, err := u.repository.FindByEmail(ctx, input.Email)
	if err != nil {
		span.AddStatus(observability.Error, "error find user by e-mail")
		span.AddAttributes(
			observability.Attributes{Key: EmailKey, Value: input.Email},
			observability.Attributes{Key: "error", Value: err.Error()},
		)
		return nil, err
	}

	if user == nil {
		span.AddStatus(observability.Error, "user not found")
		span.AddAttributes(
			observability.Attributes{Key: EmailKey, Value: input.Email},
			observability.Attributes{Key: "error", Value: err.Error()},
		)
		return nil, ErrUserNotFound
	}

	if !u.hash.CheckHash(user.Password, input.Password) {
		span.AddStatus(observability.Error, "error checking hash")
		span.AddAttributes(
			observability.Attributes{Key: EmailKey, Value: input.Email},
			observability.Attributes{Key: "error", Value: err.Error()},
		)
		return nil, ErrCheckHash
	}

	token, err := u.jwt.GenerateToken(ctx, user.ID.String(), user.Email.String())
	if err != nil {
		span.AddStatus(observability.Error, "error generate token")
		span.AddAttributes(
			observability.Attributes{Key: EmailKey, Value: input.Email},
			observability.Attributes{Key: "error", Value: err.Error()},
		)
		return nil, err
	}
	return NewAuthOutput(token, u.config.AuthExpirationAt), nil
}
