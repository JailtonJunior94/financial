package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/user/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type (
	DeleteUserUseCase interface {
		Execute(ctx context.Context, id string) error
	}

	deleteUserUseCase struct {
		o11y       observability.Observability
		fm         *metrics.FinancialMetrics
		repository interfaces.UserRepository
	}
)

func NewDeleteUserUseCase(
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
	repository interfaces.UserRepository,
) DeleteUserUseCase {
	return &deleteUserUseCase{
		o11y:       o11y,
		fm:         fm,
		repository: repository,
	}
}

func (u *deleteUserUseCase) Execute(ctx context.Context, id string) error {
	start := time.Now()
	ctx, span := u.o11y.Tracer().Start(ctx, "delete_user_usecase.execute")
	defer span.End()

	span.AddEvent("deleting user", observability.Field{Key: "user.id", Value: id})

	if err := u.repository.SoftDelete(ctx, id); err != nil {
		span.RecordError(err)
		u.fm.RecordUsecaseFailure(ctx, "delete_user", "user", "infra", time.Since(start))
		return err
	}

	u.fm.RecordUsecaseOperation(ctx, "delete_user", "user", time.Since(start))
	return nil
}
