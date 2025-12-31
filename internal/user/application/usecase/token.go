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
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
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

	span.AddEvent("generating token",
		observability.Field{Key: "user.email", Value: input.Email},
	)

	// Validate input
	if input.Email == "" || input.Password == "" {
		span.RecordError(customErrors.ErrPasswordIsRequired)
		return nil, customErrors.ErrPasswordIsRequired
	}

	user, err := u.repository.FindByEmail(ctx, input.Email)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	if user == nil {
		span.RecordError(customErrors.ErrCheckHash)
		return nil, customErrors.ErrCheckHash
	}

	if !u.hash.CheckHash(user.Password, input.Password) {
		span.RecordError(customErrors.ErrCheckHash)
		return nil, customErrors.ErrCheckHash
	}

	token, err := u.jwt.GenerateToken(ctx, user.ID.String(), user.Email.String())
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("generating token: %w", err)
	}

	return dtos.NewAuthOutput(token, u.config.AuthConfig.AuthTokenDuration), nil
}
