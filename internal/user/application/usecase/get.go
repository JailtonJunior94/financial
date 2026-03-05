package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/user/application/dtos"
	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	customerrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)


type (
	GetUserUseCase interface {
		Execute(ctx context.Context, id string) (*dtos.UserOutput, error)
	}

	getUserUseCase struct {
		o11y       observability.Observability
		fm         *metrics.FinancialMetrics
		repository interfaces.UserRepository
	}
)

func NewGetUserUseCase(
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
	repository interfaces.UserRepository,
) GetUserUseCase {
	return &getUserUseCase{
		o11y:       o11y,
		fm:         fm,
		repository: repository,
	}
}

func (u *getUserUseCase) Execute(ctx context.Context, id string) (*dtos.UserOutput, error) {
	start := time.Now()
	ctx, span := u.o11y.Tracer().Start(ctx, "get_user_usecase.execute")
	defer span.End()

	span.AddEvent("fetching user", observability.Field{Key: "user.id", Value: id})

	user, err := u.repository.FindByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		u.fm.RecordUsecaseFailure(ctx, "get_user", "user", "infra", time.Since(start))
		return nil, err
	}

	if user == nil {
		u.fm.RecordUsecaseFailure(ctx, "get_user", "user", "not_found", time.Since(start))
		return nil, customerrors.ErrUserNotFound
	}

	u.fm.RecordUsecaseOperation(ctx, "get_user", "user", time.Since(start))
	return &dtos.UserOutput{
		ID:        user.ID.String(),
		Name:      user.Name.String(),
		Email:     user.Email.String(),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt.Ptr(),
	}, nil
}
