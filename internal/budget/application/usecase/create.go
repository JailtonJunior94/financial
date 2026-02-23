package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/internal/budget/application/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/domain/factories"
	"github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"go.opentelemetry.io/otel/trace"
)

type (
	CreateBudgetUseCase interface {
		Execute(ctx context.Context, userID string, input *dtos.BudgetCreateInput) (*dtos.BudgetOutput, error)
	}

	createBudgetUseCase struct {
		uow     uow.UnitOfWork
		o11y    observability.Observability
		metrics *metrics.FinancialMetrics
	}
)

func NewCreateBudgetUseCase(
	uow uow.UnitOfWork,
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
) CreateBudgetUseCase {
	return &createBudgetUseCase{
		uow:     uow,
		o11y:    o11y,
		metrics: fm,
	}
}

func (u *createBudgetUseCase) Execute(ctx context.Context, userID string, input *dtos.BudgetCreateInput) (*dtos.BudgetOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "create_budget_usecase.execute")
	defer span.End()

	start := time.Now()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	u.o11y.Logger().Info(ctx, "execution_started",
		observability.String("operation", "CreateBudget"),
		observability.String("layer", "usecase"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", userID),
	)

	span.AddEvent("execution_started",
		observability.String("operation", "CreateBudget"),
		observability.String("layer", "usecase"),
		observability.String("entity", "budget"),
		observability.String("user_id", userID),
	)

	newBudget, err := factories.CreateBudget(userID, input)
	if err != nil {
		return nil, err
	}

	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		budgetRepository := repositories.NewBudgetRepository(tx, u.o11y, u.metrics)

		if err := budgetRepository.Insert(ctx, newBudget); err != nil {
			return err
		}

		if err := budgetRepository.InsertItems(ctx, newBudget.Items); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		u.metrics.RecordUsecaseFailure(ctx, "CreateBudget", "budget", "infra", time.Since(start))
		u.o11y.Logger().Error(ctx, "execution_failed",
			observability.String("operation", "CreateBudget"),
			observability.String("layer", "usecase"),
			observability.String("entity", "budget"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", userID),
			observability.String("error_type", "infra"),
			observability.String("error_code", "CREATE_BUDGET_FAILED"),
			observability.Error(err),
		)
		return nil, err
	}

	u.metrics.RecordUsecaseOperation(ctx, "CreateBudget", "budget", time.Since(start))
	u.o11y.Logger().Info(ctx, "execution_completed",
		observability.String("operation", "CreateBudget"),
		observability.String("layer", "usecase"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", userID),
	)

	return &dtos.BudgetOutput{
		ID:             newBudget.ID.String(),
		UserID:         newBudget.UserID.String(),
		ReferenceMonth: newBudget.ReferenceMonth.String(),
		TotalAmount:    fmt.Sprintf("%.2f", newBudget.TotalAmount.Float()),
		SpentAmount:    fmt.Sprintf("%.2f", newBudget.SpentAmount.Float()),
		PercentageUsed: fmt.Sprintf("%.3f", newBudget.PercentageUsed.Float()),
		Currency:       string(newBudget.TotalAmount.Currency()),
		CreatedAt:      newBudget.CreatedAt,
	}, nil
}
