package usecase

import (
	"context"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
	"github.com/jailtonjunior94/financial/internal/budget/domain/interfaces"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

type (
	ReplicateBudgetUseCase interface {
		Execute(ctx context.Context, repository interfaces.BudgetRepository, sourceBudget *entities.Budget) error
	}

	replicateBudgetUseCase struct {
		o11y observability.Observability
	}
)

func NewReplicateBudgetUseCase(o11y observability.Observability) ReplicateBudgetUseCase {
	return &replicateBudgetUseCase{o11y: o11y}
}

func (u *replicateBudgetUseCase) Execute(ctx context.Context, repository interfaces.BudgetRepository, sourceBudget *entities.Budget) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "replicate_budget_usecase.execute")
	defer span.End()

	u.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "ReplicateBudget"),
		observability.String("layer", "usecase"),
		observability.String("entity", "budget"),
		observability.String("user_id", sourceBudget.UserID.String()),
	)

	nextMonth := sourceBudget.ReferenceMonth.AddMonths(1)

	existing, err := repository.FindByUserIDAndReferenceMonth(ctx, sourceBudget.UserID, nextMonth)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("replicate_budget: failed to check next month: %w", err)
	}

	if existing != nil {
		u.o11y.Logger().Debug(ctx, "next_month_budget_exists_skipping_replication",
			observability.String("operation", "ReplicateBudget"),
			observability.String("layer", "usecase"),
			observability.String("entity", "budget"),
			observability.String("user_id", sourceBudget.UserID.String()),
			observability.String("next_month", nextMonth.String()),
		)
		return nil
	}

	newBudget, err := u.buildReplicatedBudget(sourceBudget, nextMonth)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("replicate_budget: failed to build replicated budget: %w", err)
	}

	if err := repository.Insert(ctx, newBudget); err != nil {
		span.RecordError(err)
		return fmt.Errorf("replicate_budget: failed to insert replicated budget: %w", err)
	}

	if err := repository.InsertItems(ctx, newBudget.Items); err != nil {
		span.RecordError(err)
		return fmt.Errorf("replicate_budget: failed to insert replicated items: %w", err)
	}

	u.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "ReplicateBudget"),
		observability.String("layer", "usecase"),
		observability.String("entity", "budget"),
		observability.String("user_id", sourceBudget.UserID.String()),
		observability.String("next_month", nextMonth.String()),
	)

	return nil
}

func (u *replicateBudgetUseCase) buildReplicatedBudget(sourceBudget *entities.Budget, nextMonth pkgVos.ReferenceMonth) (*entities.Budget, error) {
	budgetID, err := vos.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate budget ID: %w", err)
	}

	newBudget := entities.NewBudget(sourceBudget.UserID, sourceBudget.TotalAmount, nextMonth)
	newBudget.SetID(budgetID)

	newItems := make([]*entities.BudgetItem, 0, len(sourceBudget.Items))
	for _, sourceItem := range sourceBudget.Items {
		itemID, err := vos.NewUUID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate item ID: %w", err)
		}

		newItem := entities.NewBudgetItem(newBudget.ID, newBudget.TotalAmount, sourceItem.CategoryID, sourceItem.PercentageGoal)
		newItem.SetID(itemID)
		newItems = append(newItems, newItem)
	}

	if err := newBudget.AddItems(newItems); err != nil {
		return nil, fmt.Errorf("failed to add items to replicated budget: %w", err)
	}

	return newBudget, nil
}
