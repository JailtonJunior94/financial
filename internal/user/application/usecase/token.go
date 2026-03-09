package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/auth"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/encrypt"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type TokenUseCase interface {
	Execute(ctx context.Context, input *dtos.AuthInput) (*dtos.AuthOutput, error)
}

type tokenUseCase struct {
	generator  auth.TokenGenerator
	config     *configs.Config
	hash       encrypt.HashAdapter
	repository interfaces.UserRepository
	o11y       observability.Observability
	fm         *metrics.FinancialMetrics
}

func NewTokenUseCase(
	config *configs.Config,
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
	hash encrypt.HashAdapter,
	generator auth.TokenGenerator,
	repository interfaces.UserRepository,
) TokenUseCase {
	return &tokenUseCase{
		config:     config,
		hash:       hash,
		generator:  generator,
		repository: repository,
		o11y:       o11y,
		fm:         fm,
	}
}

func (u *tokenUseCase) Execute(ctx context.Context, input *dtos.AuthInput) (*dtos.AuthOutput, error) {
	start := time.Now()
	ctx, span := u.o11y.Tracer().Start(ctx, "token_usecase.execute")
	defer span.End()

	span.AddEvent("generating token")

	if input.Email == "" {
		span.RecordError(customErrors.ErrCannotBeEmpty)
		u.fm.RecordUsecaseFailure(ctx, "token", "user", "validation", time.Since(start))
		return nil, customErrors.New("email is required", customErrors.ErrCannotBeEmpty)
	}
	if input.Password == "" {
		span.RecordError(customErrors.ErrPasswordIsRequired)
		u.fm.RecordUsecaseFailure(ctx, "token", "user", "validation", time.Since(start))
		return nil, customErrors.ErrPasswordIsRequired
	}

	user, err := u.repository.FindByEmail(ctx, input.Email)
	if err != nil {
		span.RecordError(err)
		u.fm.RecordUsecaseFailure(ctx, "token", "user", "infra", time.Since(start))
		return nil, err
	}

	if user == nil {
		span.RecordError(customErrors.ErrCheckHash)
		u.fm.RecordUsecaseFailure(ctx, "token", "user", "not_found", time.Since(start))
		return nil, customErrors.ErrCheckHash
	}

	if !u.hash.CheckHash(user.Password, input.Password) {
		span.RecordError(customErrors.ErrCheckHash)
		u.fm.RecordUsecaseFailure(ctx, "token", "user", "auth", time.Since(start))
		return nil, customErrors.ErrCheckHash
	}

	token, err := u.generator.GenerateToken(ctx, user.ID.String(), user.Email.String())
	if err != nil {
		span.RecordError(err)
		u.fm.RecordUsecaseFailure(ctx, "token", "user", "infra", time.Since(start))
		return nil, fmt.Errorf("generating token: %w", err)
	}

	u.fm.RecordUsecaseOperation(ctx, "token", "user", time.Since(start))
	return dtos.NewAuthOutput(token, u.config.AuthConfig.AuthTokenDuration), nil
}
