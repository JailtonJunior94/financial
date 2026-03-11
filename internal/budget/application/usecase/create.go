package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/internal/budget/application/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/domain"
	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
	"github.com/jailtonjunior94/financial/internal/budget/domain/factories"
	"github.com/jailtonjunior94/financial/internal/budget/domain/interfaces"
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
		uow              uow.UnitOfWork
		o11y             observability.Observability
		metrics          *metrics.FinancialMetrics
		repository       interfaces.BudgetRepository
		categoryProvider interfaces.CategoryProvider
		replicateUseCase ReplicateBudgetUseCase
	}
)

func NewCreateBudgetUseCase(
	uow uow.UnitOfWork,
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
	repository interfaces.BudgetRepository,
	categoryProvider interfaces.CategoryProvider,
	replicateUseCase ReplicateBudgetUseCase,
) CreateBudgetUseCase {
	return &createBudgetUseCase{
		uow:              uow,
		o11y:             o11y,
		metrics:          fm,
		repository:       repository,
		categoryProvider: categoryProvider,
		replicateUseCase: replicateUseCase,
	}
}

func (u *createBudgetUseCase) Execute(ctx context.Context, userID string, input *dtos.BudgetCreateInput) (*dtos.BudgetOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "create_budget_usecase.execute")
	defer span.End()

	start := time.Now()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	u.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "CreateBudget"),
		observability.String("layer", "usecase"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", userID),
	)

	categoryIDs := extractCategoryIDs(input.Items)
	if err := u.categoryProvider.ValidateCategories(ctx, userID, categoryIDs); err != nil {
		span.RecordError(err)
		return nil, err
	}

	factoryItems := make([]factories.CreateBudgetItemParams, len(input.Items))
	for i, item := range input.Items {
		factoryItems[i] = factories.CreateBudgetItemParams{
			CategoryID:     item.CategoryID,
			PercentageGoal: item.PercentageGoal,
		}
	}

	newBudget, err := factories.CreateBudget(userID, &factories.CreateBudgetParams{
		ReferenceMonth: input.ReferenceMonth,
		TotalAmount:    input.TotalAmount,
		Currency:       input.Currency,
		Items:          factoryItems,
	})
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	if err := u.uow.Do(ctx, func(ctx context.Context, _ database.DBTX) error {
		return u.persistBudget(ctx, newBudget)
	}); err != nil {
		span.RecordError(err)
		u.metrics.RecordUsecaseFailure(ctx, "CreateBudget", "budget", "infra", time.Since(start))
		u.o11y.Logger().Error(ctx, "execution_failed",
			observability.String("operation", "CreateBudget"),
			observability.String("layer", "usecase"),
			observability.String("entity", "budget"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", userID),
			observability.Error(err),
		)
		return nil, err
	}

	u.metrics.RecordUsecaseOperation(ctx, "CreateBudget", "budget", time.Since(start))
	u.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "CreateBudget"),
		observability.String("layer", "usecase"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", userID),
	)

	return buildBudgetOutput(newBudget), nil
}

func (u *createBudgetUseCase) persistBudget(ctx context.Context, budget *entities.Budget) error {
	existing, err := u.repository.FindByUserIDAndReferenceMonth(ctx, budget.UserID, budget.ReferenceMonth)
	if err != nil {
		return err
	}

	if existing != nil {
		return domain.ErrBudgetAlreadyExistsForMonth
	}

	if err := u.repository.Insert(ctx, budget); err != nil {
		return err
	}

	if err := u.repository.InsertItems(ctx, budget.Items); err != nil {
		return err
	}

	return u.replicateUseCase.Execute(ctx, u.repository, budget)
}

func extractCategoryIDs(items []dtos.BudgetItemInput) []string {
	categoryIDs := make([]string, len(items))
	for i, item := range items {
		categoryIDs[i] = item.CategoryID
	}
	return categoryIDs
}

func buildBudgetOutput(budget *entities.Budget) *dtos.BudgetOutput {
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
		}
	}

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
	}
}
