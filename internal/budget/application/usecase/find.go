package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"go.opentelemetry.io/otel/trace"

	"github.com/jailtonjunior94/financial/internal/budget/application/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/domain"
	"github.com/jailtonjunior94/financial/internal/budget/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

type (
	FindBudgetUseCase interface {
		Execute(ctx context.Context, userID string, budgetID string) (*dtos.BudgetOutput, error)
	}

	findBudgetUseCase struct {
		budgetRepository interfaces.BudgetRepository
		o11y             observability.Observability
		metrics          *metrics.FinancialMetrics
	}
)

func NewFindBudgetUseCase(
	budgetRepository interfaces.BudgetRepository,
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
) FindBudgetUseCase {
	return &findBudgetUseCase{
		budgetRepository: budgetRepository,
		o11y:             o11y,
		metrics:          fm,
	}
}

func (u *findBudgetUseCase) Execute(ctx context.Context, userID string, budgetID string) (*dtos.BudgetOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "find_budget_usecase.execute")
	defer span.End()

	start := time.Now()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	u.o11y.Logger().Info(ctx, "execution_started",
		observability.String("operation", "FindBudget"),
		observability.String("layer", "usecase"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", userID),
	)

	span.AddEvent("execution_started",
		observability.String("operation", "FindBudget"),
		observability.String("layer", "usecase"),
		observability.String("entity", "budget"),
		observability.String("user_id", userID),
	)

	// Parse userID
	uid, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	// Parse budget ID
	id, err := vos.NewUUIDFromString(budgetID)
	if err != nil {
		return nil, fmt.Errorf("invalid budget_id: %w", err)
	}

	// Find budget (scoped by userID to prevent IDOR)
	budget, err := u.budgetRepository.FindByID(ctx, uid, id)
	if err != nil {
		u.metrics.RecordUsecaseFailure(ctx, "FindBudget", "budget", "infra", time.Since(start))
		u.o11y.Logger().Error(ctx, "execution_failed",
			observability.String("operation", "FindBudget"),
			observability.String("layer", "usecase"),
			observability.String("entity", "budget"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", userID),
			observability.String("error_type", "infra"),
			observability.String("error_code", "FIND_BUDGET_FAILED"),
			observability.Error(err),
		)
		return nil, err
	}

	if budget == nil {
		u.metrics.RecordUsecaseFailure(ctx, "FindBudget", "budget", "business", time.Since(start))
		u.o11y.Logger().Error(ctx, "execution_failed",
			observability.String("operation", "FindBudget"),
			observability.String("layer", "usecase"),
			observability.String("entity", "budget"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", userID),
			observability.String("error_type", "business"),
			observability.String("error_code", "BUDGET_NOT_FOUND"),
		)
		return nil, domain.ErrBudgetNotFound
	}

	u.metrics.RecordUsecaseOperation(ctx, "FindBudget", "budget", time.Since(start))
	u.o11y.Logger().Info(ctx, "execution_completed",
		observability.String("operation", "FindBudget"),
		observability.String("layer", "usecase"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", userID),
	)

	// Build items output
	items := make([]dtos.BudgetItemOutput, len(budget.Items))
	for i, item := range budget.Items {
		items[i] = dtos.BudgetItemOutput{
			ID:              item.ID.String(),
			BudgetID:        item.BudgetID.String(),
			CategoryID:      item.CategoryID.String(),
			PercentageGoal:  fmt.Sprintf("%.3f", item.PercentageGoal.Float()),
			PlannedAmount:   fmt.Sprintf("%.2f", item.PlannedAmount.Float()),
			SpentAmount:     fmt.Sprintf("%.2f", item.SpentAmount.Float()),
			RemainingAmount: fmt.Sprintf("%.2f", item.RemainingAmount().Float()),
			PercentageSpent: fmt.Sprintf("%.3f", item.PercentageSpent().Float()),
			CreatedAt:       item.CreatedAt,
			UpdatedAt:       item.UpdatedAt.ValueOr(time.Time{}),
		}
	}

	// Build budget output
	return &dtos.BudgetOutput{
		ID:             budget.ID.String(),
		UserID:         budget.UserID.String(),
		ReferenceMonth: budget.ReferenceMonth.String(),
		TotalAmount:    fmt.Sprintf("%.2f", budget.TotalAmount.Float()),
		SpentAmount:    fmt.Sprintf("%.2f", budget.SpentAmount.Float()),
		PercentageUsed: fmt.Sprintf("%.3f", budget.PercentageUsed.Float()),
		Currency:       string(budget.TotalAmount.Currency()),
		Items:          items,
		CreatedAt:      budget.CreatedAt,
		UpdatedAt:      budget.UpdatedAt.ValueOr(time.Time{}),
	}, nil
}
