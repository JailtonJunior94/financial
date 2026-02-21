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
	"github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

type (
	SyncBudgetSpentAmountUseCase interface {
		Execute(ctx context.Context, userID vos.UUID, referenceMonth pkgVos.ReferenceMonth, categoryID vos.UUID) error
	}

	syncBudgetSpentAmountUseCase struct {
		uow                  uow.UnitOfWork
		invoiceCategoryTotal interfaces.InvoiceCategoryTotalProvider
		o11y                 observability.Observability
	}
)

func NewSyncBudgetSpentAmountUseCase(
	uow uow.UnitOfWork,
	invoiceCategoryTotal interfaces.InvoiceCategoryTotalProvider,
	o11y observability.Observability,
) SyncBudgetSpentAmountUseCase {
	return &syncBudgetSpentAmountUseCase{
		uow:                  uow,
		invoiceCategoryTotal: invoiceCategoryTotal,
		o11y:                 o11y,
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

	// Leitura dos totais de fatura ANTES da transação (usa conexão raw do pool).
	// Consistente com o padrão já adotado em SyncMonthlyFromInvoicesUseCase.
	categoryTotal, err := u.invoiceCategoryTotal.GetCategoryTotal(ctx, userID, referenceMonth, categoryID)
	if err != nil {
		return fmt.Errorf("failed to get category invoice total: %w", err)
	}

	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		budgetRepository := repositories.NewBudgetRepository(tx, u.o11y)

		budget, err := budgetRepository.FindByUserIDAndReferenceMonth(ctx, userID, referenceMonth)
		if err != nil {
			return err
		}

		if budget == nil {
			u.o11y.Logger().Warn(ctx, "budget not found for user/month - ignoring event",
				observability.String("user_id", userID.String()),
				observability.String("reference_month", referenceMonth.String()),
			)
			return nil
		}

		targetItem := findItemByCategory(budget, categoryID)
		if targetItem == nil {
			u.o11y.Logger().Warn(ctx, "budget item not found for category - ignoring event",
				observability.String("budget_id", budget.ID.String()),
				observability.String("category_id", categoryID.String()),
			)
			return nil
		}

		if err := budget.UpdateItemSpentAmount(targetItem.ID, categoryTotal); err != nil {
			return err
		}

		if err := budgetRepository.UpdateItem(ctx, targetItem); err != nil {
			return err
		}

		if err := budgetRepository.Update(ctx, budget); err != nil {
			return err
		}

		u.o11y.Logger().Info(ctx, "budget spent amount synced",
			observability.String("budget_id", budget.ID.String()),
			observability.String("item_id", targetItem.ID.String()),
			observability.String("category_id", categoryID.String()),
			observability.Int64("total_cents", categoryTotal.Cents()),
		)

		return nil
	})

	if err != nil {
		u.o11y.Logger().Error(ctx, "failed to sync budget spent amount", observability.Error(err))
		return err
	}

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
