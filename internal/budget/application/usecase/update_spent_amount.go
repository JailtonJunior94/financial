package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financial/internal/budget/application/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/domain"
	"github.com/jailtonjunior94/financial/internal/budget/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/money"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	UpdateSpentAmountUseCase interface {
		Execute(ctx context.Context, userID, budgetID, itemID string, input *dtos.UpdateSpentAmountInput) error
	}

	updateSpentAmountUseCase struct {
		uow         uow.UnitOfWork
		o11y        observability.Observability
		metrics     *metrics.FinancialMetrics
		repoFactory interfaces.BudgetRepositoryFactory
	}
)

func NewUpdateSpentAmountUseCase(
	uow uow.UnitOfWork,
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
	repoFactory interfaces.BudgetRepositoryFactory,
) UpdateSpentAmountUseCase {
	return &updateSpentAmountUseCase{
		uow:         uow,
		o11y:        o11y,
		metrics:     fm,
		repoFactory: repoFactory,
	}
}

func (u *updateSpentAmountUseCase) Execute(ctx context.Context, userID, budgetID, itemID string, input *dtos.UpdateSpentAmountInput) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "update_spent_amount_usecase.execute")
	defer span.End()

	// Parse userID
	userUUID, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Parse budgetID
	budgetUUID, err := vos.NewUUIDFromString(budgetID)
	if err != nil {
		return fmt.Errorf("invalid budget ID: %w", err)
	}

	// Parse itemID
	itemUUID, err := vos.NewUUIDFromString(itemID)
	if err != nil {
		return fmt.Errorf("invalid item ID: %w", err)
	}

	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		budgetRepository := u.repoFactory(tx)

		// Buscar o orçamento completo (scoped by userID to prevent IDOR)
		budget, err := budgetRepository.FindByID(ctx, userUUID, budgetUUID)
		if err != nil {
			return err
		}
		if budget == nil {
			return domain.ErrBudgetNotFound
		}

		// Encontrar o item
		item := budget.FindItemByID(itemUUID)
		if item == nil {
			return domain.ErrBudgetItemNotFound
		}

		// Parse spent amount from string (half-even rounding)
		spentMoney, err := money.NewMoney(input.SpentAmount, budget.TotalAmount.Currency())
		if err != nil {
			return fmt.Errorf("invalid spent amount: %w", err)
		}

		// Atualizar o gasto através do aggregate root (aplica validações de negócio)
		if err := budget.UpdateItemSpentAmount(itemUUID, spentMoney); err != nil {
			return err
		}

		// Persistir as alterações
		if err := budgetRepository.UpdateItem(ctx, item); err != nil {
			return err
		}

		// Atualizar os totais do budget
		if err := budgetRepository.Update(ctx, budget); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(observability.StatusCodeError, "failed to update spent amount")
		u.o11y.Logger().Error(ctx, "failed to update spent amount",
			observability.String("operation", "update_spent_amount"),
			observability.String("layer", "usecase"),
			observability.String("entity", "budget"),
			observability.Error(err))
		return err
	}

	return nil
}
