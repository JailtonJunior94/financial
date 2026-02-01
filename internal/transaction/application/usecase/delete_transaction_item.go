package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	DeleteTransactionItemUseCase interface {
		Execute(ctx context.Context, userID, itemID string) (*dtos.MonthlyTransactionOutput, error)
	}

	deleteTransactionItemUseCase struct {
		uow  uow.UnitOfWork
		repo interfaces.TransactionRepository
		o11y observability.Observability
	}
)

func NewDeleteTransactionItemUseCase(
	uow uow.UnitOfWork,
	repo interfaces.TransactionRepository,
	o11y observability.Observability,
) DeleteTransactionItemUseCase {
	return &deleteTransactionItemUseCase{
		uow:  uow,
		repo: repo,
		o11y: o11y,
	}
}

func (u *deleteTransactionItemUseCase) Execute(
	ctx context.Context,
	userID, itemID string,
) (*dtos.MonthlyTransactionOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "delete_transaction_item_usecase.execute")
	defer span.End()

	// Parse IDs
	user, err := sharedVos.NewUUIDFromString(userID)
	if err != nil {
		u.o11y.Logger().Error(ctx, "invalid user id", observability.Error(err))
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	itemUUID, err := sharedVos.NewUUIDFromString(itemID)
	if err != nil {
		u.o11y.Logger().Error(ctx, "invalid item id", observability.Error(err))
		return nil, fmt.Errorf("invalid item ID: %w", err)
	}

	// Execute within transaction
	var monthly *dtos.MonthlyTransactionOutput
	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		// Find item
		item, err := u.repo.FindItemByID(ctx, tx, user, itemUUID)
		if err != nil {
			return fmt.Errorf("failed to find item: %w", err)
		}

		if item == nil {
			return fmt.Errorf("item not found")
		}

		// Find monthly aggregate
		monthlyAggregate, err := u.repo.FindMonthlyByID(ctx, tx, user, item.MonthlyID)
		if err != nil {
			return fmt.Errorf("failed to find monthly transaction: %w", err)
		}

		// Remove item through aggregate (soft delete + recalculates totals)
		if err := monthlyAggregate.RemoveItem(itemUUID); err != nil {
			return fmt.Errorf("failed to remove item: %w", err)
		}

		// Persist soft delete
		if err := u.repo.UpdateItem(ctx, tx, item); err != nil {
			return fmt.Errorf("failed to persist item deletion: %w", err)
		}

		// Update aggregate totals
		if err := u.repo.UpdateMonthly(ctx, tx, monthlyAggregate); err != nil {
			return fmt.Errorf("failed to update monthly totals: %w", err)
		}

		// Convert to DTO
		monthly = u.toOutput(monthlyAggregate)
		return nil
	})

	if err != nil {
		u.o11y.Logger().Error(ctx, "failed to delete transaction item", observability.Error(err))
		return nil, err
	}

	return monthly, nil
}

func (u *deleteTransactionItemUseCase) toOutput(aggregate any) *dtos.MonthlyTransactionOutput {
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
