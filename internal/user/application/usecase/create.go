package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/domain/factories"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/encrypt"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type (
	CreateUserUseCase interface {
		Execute(ctx context.Context, input *dtos.CreateUserInput) (*dtos.CreateUserOutput, error)
	}

	createUserUseCase struct {
		o11y       observability.Observability
		fm         *metrics.FinancialMetrics
		hash       encrypt.HashAdapter
		repository interfaces.UserRepository
	}
)

func NewCreateUserUseCase(
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
	hash encrypt.HashAdapter,
	repository interfaces.UserRepository,
) CreateUserUseCase {
	return &createUserUseCase{
		o11y:       o11y,
		fm:         fm,
		hash:       hash,
		repository: repository,
	}
}

func (u *createUserUseCase) Execute(ctx context.Context, input *dtos.CreateUserInput) (*dtos.CreateUserOutput, error) {
	start := time.Now()
	ctx, span := u.o11y.Tracer().Start(ctx, "create_user_usecase.execute")
	defer span.End()

	span.AddEvent("creating user",
		observability.Field{Key: "user.email", Value: input.Email},
		observability.Field{Key: "user.name", Value: input.Name},
	)

	if input.Password == "" {
		span.RecordError(customErrors.ErrPasswordIsRequired)
		u.fm.RecordUsecaseFailure(ctx, "create_user", "user", "validation", time.Since(start))
		return nil, customErrors.ErrPasswordIsRequired
	}

	user, err := factories.CreateUser(input.Name, input.Email)
	if err != nil {
		span.RecordError(err)
		u.fm.RecordUsecaseFailure(ctx, "create_user", "user", "validation", time.Since(start))
		return nil, err
	}

	hash, err := u.hash.GenerateHash(input.Password)
	if err != nil {
		span.RecordError(err)
		u.fm.RecordUsecaseFailure(ctx, "create_user", "user", "infra", time.Since(start))
		return nil, fmt.Errorf("generating password hash: %w", err)
	}

	if err := user.SetPassword(hash); err != nil {
		span.RecordError(err)
		u.fm.RecordUsecaseFailure(ctx, "create_user", "user", "validation", time.Since(start))
		return nil, err
	}

	userCreated, err := u.repository.Insert(ctx, user)
	if err != nil {
		span.RecordError(err)
		u.fm.RecordUsecaseFailure(ctx, "create_user", "user", "infra", time.Since(start))
		return nil, err
	}

	u.fm.RecordUsecaseOperation(ctx, "create_user", "user", time.Since(start))
	return &dtos.CreateUserOutput{
		ID:        userCreated.ID.String(),
		Name:      userCreated.Name.String(),
		Email:     userCreated.Email.String(),
		CreatedAt: userCreated.CreatedAt,
	}, nil
}
