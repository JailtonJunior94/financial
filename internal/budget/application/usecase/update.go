package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/internal/budget/application/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/domain"
	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
	"github.com/jailtonjunior94/financial/internal/budget/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/money"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"go.opentelemetry.io/otel/trace"
)

type (
	UpdateBudgetUseCase interface {
		Execute(ctx context.Context, userID string, budgetID string, input *dtos.BudgetUpdateInput) (*dtos.BudgetOutput, error)
	}

	updateBudgetUseCase struct {
		uow              uow.UnitOfWork
		o11y             observability.Observability
		metrics          *metrics.FinancialMetrics
		repository       interfaces.BudgetRepository
		categoryProvider interfaces.CategoryProvider
		replicateUseCase ReplicateBudgetUseCase
	}
)

func NewUpdateBudgetUseCase(
	uow uow.UnitOfWork,
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
	repository interfaces.BudgetRepository,
	categoryProvider interfaces.CategoryProvider,
	replicateUseCase ReplicateBudgetUseCase,
) UpdateBudgetUseCase {
	return &updateBudgetUseCase{
		uow:              uow,
		o11y:             o11y,
		metrics:          fm,
		repository:       repository,
		categoryProvider: categoryProvider,
		replicateUseCase: replicateUseCase,
	}
}

func (u *updateBudgetUseCase) Execute(ctx context.Context, userID string, budgetID string, input *dtos.BudgetUpdateInput) (*dtos.BudgetOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "update_budget_usecase.execute")
	defer span.End()

	start := time.Now()
	correlationID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

	u.o11y.Logger().Info(ctx, "request_received",
		observability.String("operation", "UpdateBudget"),
		observability.String("layer", "usecase"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", userID),
	)

	uid, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	id, err := vos.NewUUIDFromString(budgetID)
	if err != nil {
		return nil, fmt.Errorf("invalid budget_id: %w", err)
	}

	categoryIDs := extractCategoryIDs(input.Items)
	if err := u.categoryProvider.ValidateCategories(ctx, userID, categoryIDs); err != nil {
		span.RecordError(err)
		return nil, err
	}

	var updatedBudget *dtos.BudgetOutput
	if err := u.uow.Do(ctx, func(ctx context.Context, _ database.DBTX) error {
		result, err := u.performUpdate(ctx, uid, id, input)
		if err != nil {
			return err
		}
		updatedBudget = result
		return nil
	}); err != nil {
		span.RecordError(err)
		u.metrics.RecordUsecaseFailure(ctx, "UpdateBudget", "budget", "infra", time.Since(start))
		u.o11y.Logger().Error(ctx, "execution_failed",
			observability.String("operation", "UpdateBudget"),
			observability.String("layer", "usecase"),
			observability.String("entity", "budget"),
			observability.String("correlation_id", correlationID),
			observability.String("user_id", userID),
			observability.Error(err),
		)
		return nil, err
	}

	u.metrics.RecordUsecaseOperation(ctx, "UpdateBudget", "budget", time.Since(start))
	u.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "UpdateBudget"),
		observability.String("layer", "usecase"),
		observability.String("entity", "budget"),
		observability.String("correlation_id", correlationID),
		observability.String("user_id", userID),
	)

	return updatedBudget, nil
}

func (u *updateBudgetUseCase) performUpdate(ctx context.Context, uid, id vos.UUID, input *dtos.BudgetUpdateInput) (*dtos.BudgetOutput, error) {
	budget, err := u.repository.FindByID(ctx, uid, id)
	if err != nil {
		return nil, err
	}

	if budget == nil {
		return nil, domain.ErrBudgetNotFound
	}

	newTotalAmount, err := money.NewMoney(input.TotalAmount, budget.TotalAmount.Currency())
	if err != nil {
		return nil, fmt.Errorf("invalid total_amount: %w", err)
	}

	if newTotalAmount.IsNegative() || newTotalAmount.IsZero() {
		return nil, fmt.Errorf("total_amount must be positive: %w", domain.ErrNegativeAmount)
	}

	budget.TotalAmount = newTotalAmount
	budget.UpdatedAt = vos.NewNullableTime(time.Now().UTC())

	existingItems, newItems, err := u.buildUpdatedItems(budget, input.Items, newTotalAmount)
	if err != nil {
		return nil, err
	}

	allItems := make([]*entities.BudgetItem, 0, len(existingItems)+len(newItems))
	allItems = append(allItems, existingItems...)
	allItems = append(allItems, newItems...)
	budget.Items = []*entities.BudgetItem{}
	if err := budget.AddItems(allItems); err != nil {
		return nil, err
	}
	if err := budget.RecalculateTotals(); err != nil {
		return nil, err
	}

	if err := u.persistBudgetUpdate(ctx, budget, existingItems, newItems); err != nil {
		return nil, err
	}

	if err := u.replicateUseCase.Execute(ctx, u.repository, budget); err != nil {
		return nil, err
	}

	return buildBudgetOutput(budget), nil
}

func (u *updateBudgetUseCase) buildUpdatedItems(budget *entities.Budget, inputItems []dtos.BudgetItemInput, newTotalAmount vos.Money) (existingItems, newItems []*entities.BudgetItem, err error) {
	seenCategories := make(map[string]bool)
	for _, item := range inputItems {
		if seenCategories[item.CategoryID] {
			return nil, nil, domain.ErrDuplicateCategory
		}
		seenCategories[item.CategoryID] = true
	}

	existingByCategory := make(map[string]*entities.BudgetItem)
	for _, item := range budget.Items {
		existingByCategory[item.CategoryID.String()] = item
	}

	for _, inputItem := range inputItems {
		categoryID, err := vos.NewUUIDFromString(inputItem.CategoryID)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid category_id %s: %w", inputItem.CategoryID, err)
		}

		percentage, err := money.NewPercentageFromString(inputItem.PercentageGoal)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid percentage_goal for category %s: %w", inputItem.CategoryID, err)
		}

		if existing, ok := existingByCategory[inputItem.CategoryID]; ok {
			plannedAmount, err := percentage.Apply(newTotalAmount)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to calculate planned_amount: %w", err)
			}
			existing.PercentageGoal = percentage
			existing.PlannedAmount = plannedAmount
			existing.UpdatedAt = vos.NewNullableTime(time.Now().UTC())
			existingItems = append(existingItems, existing)
		} else {
			itemID, err := vos.NewUUID()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to generate item ID: %w", err)
			}
			newItem := entities.NewBudgetItem(budget.ID, newTotalAmount, categoryID, percentage)
			newItem.SetID(itemID)
			newItems = append(newItems, newItem)
		}
	}

	return existingItems, newItems, nil
}

func (u *updateBudgetUseCase) persistBudgetUpdate(ctx context.Context, budget *entities.Budget, existingItems, newItems []*entities.BudgetItem) error {
	if err := u.repository.Update(ctx, budget); err != nil {
		return err
	}

	for _, item := range existingItems {
		if err := u.repository.UpdateItem(ctx, item); err != nil {
			return err
		}
	}

	if len(newItems) > 0 {
		if err := u.repository.InsertItems(ctx, newItems); err != nil {
			return err
		}
	}

	keepIDs := make([]vos.UUID, 0, len(existingItems)+len(newItems))
	for _, item := range existingItems {
		keepIDs = append(keepIDs, item.ID)
	}
	for _, item := range newItems {
		keepIDs = append(keepIDs, item.ID)
	}

	return u.repository.DeleteItemsNotIn(ctx, budget.ID, keepIDs)
}
