package usecase

import (
	"context"
	"errors"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/authentication"
	"github.com/jailtonjunior94/financial/pkg/encrypt"
	"github.com/jailtonjunior94/financial/pkg/logger"

	"go.opentelemetry.io/otel/trace"
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
	trace      trace.Tracer
	config     *configs.Config
	logger     logger.Logger
	hash       encrypt.HashAdapter
	jwt        authentication.JwtAdapter
	repository interfaces.UserRepository
}

func NewTokenUseCase(
	trace trace.Tracer,
	config *configs.Config,
	logger logger.Logger,
	hash encrypt.HashAdapter,
	jwt authentication.JwtAdapter,
	repository interfaces.UserRepository,
) TokenUseCase {
	return &tokenUseCase{
		trace:      trace,
		config:     config,
		logger:     logger,
		hash:       hash,
		jwt:        jwt,
		repository: repository,
	}
}

func (u *tokenUseCase) Execute(ctx context.Context, input *AuthInput) (*AuthOutput, error) {
	ctx, span := u.trace.Start(ctx, "token")
	defer span.End()

	user, err := u.repository.FindByEmail(ctx, input.Email)
	if err != nil {
		u.logger.Error("error find user by e-mail",
			logger.Field{Key: EmailKey, Value: input.Email},
			logger.Field{Key: ErrorKey, Value: err.Error()},
		)
		return nil, err
	}

	if user == nil {
		u.logger.Warn("user not found",
			logger.Field{Key: EmailKey, Value: input.Email},
			logger.Field{Key: ErrorKey, Value: err.Error()},
		)
		return nil, ErrUserNotFound
	}

	if !u.hash.CheckHash(user.Password, input.Password) {
		u.logger.Warn("error checking hash",
			logger.Field{Key: EmailKey, Value: input.Email},
		)
		return nil, ErrCheckHash
	}

	token, err := u.jwt.GenerateToken(user.ID, user.Email)
	if err != nil {
		u.logger.Error("error generate token",
			logger.Field{Key: EmailKey, Value: input.Email},
			logger.Field{Key: ErrorKey, Value: err.Error()},
		)
		return nil, err
	}
	return NewAuthOutput(token, u.config.AuthExpirationAt), nil
}
