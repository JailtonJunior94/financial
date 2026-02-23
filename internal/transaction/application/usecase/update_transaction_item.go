package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/strategies"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/money"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	UpdateTransactionItemUseCase interface {
		Execute(ctx context.Context, userID, itemID string, input *dtos.UpdateTransactionItemInput) (*dtos.MonthlyTransactionOutput, error)
	}

	updateTransactionItemUseCase struct {
		uow  uow.UnitOfWork
		repo interfaces.TransactionRepository
		o11y observability.Observability
		fm   *metrics.FinancialMetrics
	}
)

func NewUpdateTransactionItemUseCase(
	uow uow.UnitOfWork,
	repo interfaces.TransactionRepository,
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
) UpdateTransactionItemUseCase {
	return &updateTransactionItemUseCase{
		uow:  uow,
		repo: repo,
		o11y: o11y,
		fm:   fm,
	}
}

func (u *updateTransactionItemUseCase) Execute(
	ctx context.Context,
	userID, itemID string,
	input *dtos.UpdateTransactionItemInput,
) (*dtos.MonthlyTransactionOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "update_transaction_item_usecase.execute")
	defer span.End()

	// Parse IDs
	user, err := sharedVos.NewUUIDFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	itemUUID, err := sharedVos.NewUUIDFromString(itemID)
	if err != nil {
		return nil, fmt.Errorf("invalid item ID: %w", err)
	}

	// Parse input
	direction, err := transactionVos.NewTransactionDirection(input.Direction)
	if err != nil {
		return nil, fmt.Errorf("invalid direction: %w", err)
	}

	transactionType, err := transactionVos.NewTransactionType(input.Type)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction type: %w", err)
	}

	amount, err := money.NewMoneyBRL(input.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	// Validate using strategy
	strategy := strategies.GetStrategy(transactionType)
	if strategy == nil {
		return nil, fmt.Errorf("unsupported transaction type: %s", input.Type)
	}

	if err := strategy.Validate(amount, direction, input.IsPaid); err != nil {
		return nil, err
	}

	// Execute within transaction
	var monthly *dtos.MonthlyTransactionOutput
	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		// Busca o item para obter o MonthlyID do aggregate
		item, err := u.repo.FindItemByID(ctx, tx, user, itemUUID)
		if err != nil {
			return fmt.Errorf("failed to find item: %w", err)
		}
		if item == nil {
			return fmt.Errorf("transaction item not found")
		}

		// Carrega o aggregate completo
		monthlyAggregate, err := u.repo.FindMonthlyByID(ctx, tx, user, item.MonthlyID)
		if err != nil {
			return fmt.Errorf("failed to find monthly transaction: %w", err)
		}
		if monthlyAggregate == nil {
			return fmt.Errorf("monthly transaction not found")
		}

		// Aplica a mutação via aggregate (recalcula totais internamente)
		if err := monthlyAggregate.UpdateItem(
			itemUUID,
			input.Title,
			input.Description,
			amount,
			direction,
			transactionType,
			input.IsPaid,
		); err != nil {
			return fmt.Errorf("failed to update item: %w", err)
		}

		// Obtém o item ATUALIZADO do aggregate (não a cópia antiga de FindItemByID)
		updatedItem := monthlyAggregate.FindItemByID(itemUUID)
		if updatedItem == nil {
			return fmt.Errorf("item not found in aggregate after update")
		}

		// Persiste o item com os valores atualizados
		if err := u.repo.UpdateItem(ctx, tx, updatedItem); err != nil {
			return fmt.Errorf("failed to persist item: %w", err)
		}

		if err := u.repo.UpdateMonthly(ctx, tx, monthlyAggregate); err != nil {
			return fmt.Errorf("failed to update monthly totals: %w", err)
		}

		monthly = monthlyToOutput(monthlyAggregate)
		return nil
	})

	if err != nil {
		u.o11y.Logger().Error(ctx, "failed to update transaction item", observability.Error(err))
		return nil, err
	}

	return monthly, nil
}
