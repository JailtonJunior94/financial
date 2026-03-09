package usecase

import (
	"context"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
	"github.com/jailtonjunior94/financial/internal/budget/domain/interfaces"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

type (
	SyncBudgetSpentAmountUseCase interface {
		Execute(ctx context.Context, userID vos.UUID, referenceMonth pkgVos.ReferenceMonth, categoryID vos.UUID) error
	}

	syncBudgetSpentAmountUseCase struct {
		uow                  uow.UnitOfWork
		invoiceCategoryTotal interfaces.InvoiceCategoryTotalProvider
		budgetRepository     interfaces.BudgetRepository
		o11y                 observability.Observability
		fm                   *metrics.FinancialMetrics
	}
)

func NewSyncBudgetSpentAmountUseCase(
	uow uow.UnitOfWork,
	invoiceCategoryTotal interfaces.InvoiceCategoryTotalProvider,
	budgetRepository interfaces.BudgetRepository,
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
) SyncBudgetSpentAmountUseCase {
	return &syncBudgetSpentAmountUseCase{
		uow:                  uow,
		invoiceCategoryTotal: invoiceCategoryTotal,
		budgetRepository:     budgetRepository,
		o11y:                 o11y,
		fm:                   fm,
	}
}

func (u *syncBudgetSpentAmountUseCase) Execute(
	ctx context.Context,
	userID vos.UUID,
	referenceMonth pkgVos.ReferenceMonth,
	categoryID vos.UUID,
) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "sync_budget_spent_amount_usecase.execute")
	defer span.End()

	categoryTotal, err := u.invoiceCategoryTotal.GetCategoryTotal(ctx, userID, referenceMonth, categoryID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get category invoice total: %w", err)
	}

	if err := u.uow.Do(ctx, func(ctx context.Context, _ database.DBTX) error {
		return u.syncItem(ctx, userID, referenceMonth, categoryID, categoryTotal)
	}); err != nil {
		span.RecordError(err)
		u.o11y.Logger().Error(ctx, "execution_failed",
			observability.String("operation", "SyncBudgetSpentAmount"),
			observability.String("layer", "usecase"),
			observability.String("entity", "budget"),
			observability.String("user_id", userID.String()),
			observability.Error(err),
		)
		return err
	}

	return nil
}

func (u *syncBudgetSpentAmountUseCase) syncItem(
	ctx context.Context,
	userID vos.UUID,
	referenceMonth pkgVos.ReferenceMonth,
	categoryID vos.UUID,
	categoryTotal vos.Money,
) error {
	budget, err := u.budgetRepository.FindByUserIDAndReferenceMonth(ctx, userID, referenceMonth)
	if err != nil {
		return err
	}

	if budget == nil {
		u.o11y.Logger().Warn(ctx, "budget_not_found_ignoring_event",
			observability.String("user_id", userID.String()),
			observability.String("reference_month", referenceMonth.String()),
		)
		return nil
	}

	targetItem := findItemByCategory(budget, categoryID)
	if targetItem == nil {
		u.o11y.Logger().Warn(ctx, "budget_item_not_found_ignoring_event",
			observability.String("budget_id", budget.ID.String()),
			observability.String("category_id", categoryID.String()),
		)
		return nil
	}

	if err := budget.UpdateItemSpentAmount(targetItem.ID, categoryTotal); err != nil {
		return err
	}

	if err := u.budgetRepository.UpdateItem(ctx, targetItem); err != nil {
		return err
	}

	if err := u.budgetRepository.Update(ctx, budget); err != nil {
		return err
	}

	u.o11y.Logger().Info(ctx, "budget_spent_amount_synced",
		observability.String("budget_id", budget.ID.String()),
		observability.String("item_id", targetItem.ID.String()),
		observability.String("category_id", categoryID.String()),
		observability.Int64("total_cents", categoryTotal.Cents()),
	)

	return nil
}

func findItemByCategory(budget *entities.Budget, categoryID vos.UUID) *entities.BudgetItem {
	for _, item := range budget.Items {
		if item.CategoryID.String() == categoryID.String() {
			return item
		}
	}
	return nil
}
