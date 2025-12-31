package usecase

import (
	"context"
	"fmt"
	"strconv"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/budget/domain"
	"github.com/jailtonjunior94/financial/internal/budget/application/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/database/uow"
	pkgDatabase "github.com/jailtonjunior94/financial/pkg/database"
)

type (
	UpdateSpentAmountUseCase interface {
		Execute(ctx context.Context, budgetID, itemID string, input *dtos.UpdateSpentAmountInput) error
	}

	updateSpentAmountUseCase struct {
		uow  uow.UnitOfWork
		o11y observability.Observability
	}
)

func NewUpdateSpentAmountUseCase(
	uow uow.UnitOfWork,
	o11y observability.Observability,
) UpdateSpentAmountUseCase {
	return &updateSpentAmountUseCase{
		uow:  uow,
		o11y: o11y,
	}
}

func (u *updateSpentAmountUseCase) Execute(ctx context.Context, budgetID, itemID string, input *dtos.UpdateSpentAmountInput) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "update_spent_amount_usecase.execute")
	defer span.End()

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

	// Parse spent amount from string
	spentFloat, err := strconv.ParseFloat(input.SpentAmount, 64)
	if err != nil {
		return fmt.Errorf("invalid spent amount format: %w", err)
	}

	err = u.uow.Do(ctx, func(ctx context.Context, tx pkgDatabase.DBExecutor) error {
		// Criar repositório com a transação
		budgetRepository := repositories.NewBudgetRepository(tx, u.o11y)

		// Buscar o orçamento completo
		budget, err := budgetRepository.FindByID(ctx, budgetUUID)
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

		// Criar Money a partir do float (usando a moeda do orçamento)
		spentMoney, err := vos.NewMoneyFromFloat(spentFloat, budget.TotalAmount.Currency())
		if err != nil {
			return fmt.Errorf("failed to create money value: %w", err)
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
		u.o11y.Logger().Error(ctx, "failed to update spent amount", observability.Error(err))
		return err
	}

	return nil
}
