package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/strategies"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"

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
	}
)

func NewUpdateTransactionItemUseCase(
	uow uow.UnitOfWork,
	repo interfaces.TransactionRepository,
	o11y observability.Observability,
) UpdateTransactionItemUseCase {
	return &updateTransactionItemUseCase{
		uow:  uow,
		repo: repo,
		o11y: o11y,
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

	amount, err := sharedVos.NewMoneyFromString(input.Amount, sharedVos.CurrencyBRL)
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
		// Find item
		item, err := u.repo.FindItemByID(ctx, tx, user, itemUUID)
		if err != nil {
			return fmt.Errorf("failed to find item: %w", err)
		}

		// Find monthly aggregate
		monthlyAggregate, err := u.repo.FindMonthlyByID(ctx, tx, user, item.MonthlyID)
		if err != nil {
			return fmt.Errorf("failed to find monthly transaction: %w", err)
		}

		// Update item through aggregate (recalculates totals)
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

		// Persist changes
		if err := u.repo.UpdateItem(ctx, tx, item); err != nil {
			return fmt.Errorf("failed to persist item: %w", err)
		}

		if err := u.repo.UpdateMonthly(ctx, tx, monthlyAggregate); err != nil {
			return fmt.Errorf("failed to update monthly totals: %w", err)
		}

		monthly = u.toOutput(monthlyAggregate)
		return nil
	})

	if err != nil {
		u.o11y.Logger().Error(ctx, "failed to update transaction item", observability.Error(err))
		return nil, err
	}

	return monthly, nil
}

func (u *updateTransactionItemUseCase) toOutput(aggregate any) *dtos.MonthlyTransactionOutput {
	monthly, ok := aggregate.(*entities.MonthlyTransaction)
	if !ok {
		return &dtos.MonthlyTransactionOutput{}
	}

	items := make([]*dtos.TransactionItemOutput, 0)
	for _, item := range monthly.Items {
		if item.IsDeleted() {
			continue
		}
		items = append(items, &dtos.TransactionItemOutput{
			ID:          item.ID.String(),
			CategoryID:  item.CategoryID.String(),
			Title:       item.Title,
			Description: item.Description,
			Amount:      item.Amount.String(),
			Direction:   item.Direction.String(),
			Type:        item.Type.String(),
			IsPaid:      item.IsPaid,
			CreatedAt:   item.CreatedAt.ValueOr(time.Time{}),
			UpdatedAt:   item.UpdatedAt.ValueOr(time.Time{}),
		})
	}

	return &dtos.MonthlyTransactionOutput{
		ID:             monthly.ID.String(),
		ReferenceMonth: monthly.ReferenceMonth.String(),
		TotalIncome:    monthly.TotalIncome.String(),
		TotalExpense:   monthly.TotalExpense.String(),
		TotalAmount:    monthly.TotalAmount.String(),
		Items:          items,
		CreatedAt:      monthly.CreatedAt.ValueOr(time.Time{}),
		UpdatedAt:      monthly.UpdatedAt.ValueOr(time.Time{}),
	}
}
