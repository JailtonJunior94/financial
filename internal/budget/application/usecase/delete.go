package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/internal/budget/domain"
	"github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"go.opentelemetry.io/otel/trace"
)

type (
	DeleteBudgetUseCase interface {
		Execute(ctx context.Context, userID string, budgetID string) error
	}

	deleteBudgetUseCase struct {
		uow     uow.UnitOfWork
		o11y    observability.Observability
		metrics *metrics.FinancialMetrics
	}
)

func NewDeleteBudgetUseCase(
	uow uow.UnitOfWork,
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
) DeleteBudgetUseCase {
	return &deleteBudgetUseCase{
		uow:     uow,
		o11y:    o11y,
		metrics: fm,
	}
}

func (u *deleteBudgetUseCase) Execute(ctx context.Context, userID string, budgetID string) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "delete_budget_usecase.execute")
	defer span.End()

	start := time.Now()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	u.o11y.Logger().Info(ctx, "execution_started",
		observability.String("operation", "DeleteBudget"),
		observability.String("layer", "usecase"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", userID),
	)

	span.AddEvent("execution_started",
		observability.String("operation", "DeleteBudget"),
		observability.String("layer", "usecase"),
		observability.String("entity", "budget"),
		observability.String("user_id", userID),
	)

	// Parse userID
	uid, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	// Parse budget ID
	id, err := vos.NewUUIDFromString(budgetID)
	if err != nil {
		return fmt.Errorf("invalid budget_id: %w", err)
	}

	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		budgetRepository := repositories.NewBudgetRepository(tx, u.o11y, u.metrics)

		// Find budget (scoped by userID to prevent IDOR)
		budget, err := budgetRepository.FindByID(ctx, uid, id)
		if err != nil {
			return err
		}

		if budget == nil {
			return domain.ErrBudgetNotFound
		}

		// Soft delete via repositorio (persiste deleted_at no banco)
		if err := budgetRepository.Delete(ctx, id); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		u.metrics.RecordUsecaseFailure(ctx, "DeleteBudget", "budget", "infra", time.Since(start))
		u.o11y.Logger().Error(ctx, "execution_failed",
			observability.String("operation", "DeleteBudget"),
			observability.String("layer", "usecase"),
			observability.String("entity", "budget"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", userID),
			observability.String("error_type", "infra"),
			observability.String("error_code", "DELETE_BUDGET_FAILED"),
			observability.Error(err),
		)
		return err
	}

	u.metrics.RecordUsecaseOperation(ctx, "DeleteBudget", "budget", time.Since(start))
	u.o11y.Logger().Info(ctx, "execution_completed",
		observability.String("operation", "DeleteBudget"),
		observability.String("layer", "usecase"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", userID),
	)

	return nil
}
