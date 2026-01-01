package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/strategies"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/database/uow"
	pkgDatabase "github.com/jailtonjunior94/financial/pkg/database"
)

type (
	RegisterTransactionUseCase interface {
		Execute(ctx context.Context, userID string, input *dtos.RegisterTransactionInput) (*dtos.MonthlyTransactionOutput, error)
	}

	registerTransactionUseCase struct {
		uow  uow.UnitOfWork
		repo interfaces.TransactionRepository
		o11y observability.Observability
	}
)

func NewRegisterTransactionUseCase(
	uow uow.UnitOfWork,
	repo interfaces.TransactionRepository,
	o11y observability.Observability,
) RegisterTransactionUseCase {
	return &registerTransactionUseCase{
		uow:  uow,
		repo: repo,
		o11y: o11y,
	}
}

func (u *registerTransactionUseCase) Execute(
	ctx context.Context,
	userID string,
	input *dtos.RegisterTransactionInput,
) (*dtos.MonthlyTransactionOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "register_transaction_usecase.execute")
	defer span.End()

	// Parse userID
	user, err := sharedVos.NewUUIDFromString(userID)
	if err != nil {
		u.o11y.Logger().Error(ctx, "invalid user id", observability.Error(err))
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Parse categoryID
	categoryID, err := sharedVos.NewUUIDFromString(input.CategoryID)
	if err != nil {
		u.o11y.Logger().Error(ctx, "invalid category id", observability.Error(err))
		return nil, fmt.Errorf("invalid category ID: %w", err)
	}

	// Parse reference month
	referenceMonth, err := transactionVos.NewReferenceMonthFromString(input.ReferenceMonth)
	if err != nil {
		u.o11y.Logger().Error(ctx, "invalid reference month", observability.Error(err))
		return nil, fmt.Errorf("invalid reference month: %w", err)
	}

	// Parse direction
	direction, err := transactionVos.NewTransactionDirection(input.Direction)
	if err != nil {
		u.o11y.Logger().Error(ctx, "invalid direction", observability.Error(err))
		return nil, fmt.Errorf("invalid direction: %w", err)
	}

	// Parse transaction type
	transactionType, err := transactionVos.NewTransactionType(input.Type)
	if err != nil {
		u.o11y.Logger().Error(ctx, "invalid transaction type", observability.Error(err))
		return nil, fmt.Errorf("invalid transaction type: %w", err)
	}

	// Parse amount (Money VO)
	amount, err := sharedVos.NewMoneyFromFloat(input.Amount, sharedVos.CurrencyBRL)
	if err != nil {
		u.o11y.Logger().Error(ctx, "invalid amount", observability.Error(err))
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	// Get strategy for validation and creation
	strategy := strategies.GetStrategy(transactionType)
	if strategy == nil {
		u.o11y.Logger().Error(ctx, "strategy not found", observability.String("type", input.Type))
		return nil, fmt.Errorf("unsupported transaction type: %s", input.Type)
	}

	// Execute within transaction
	var monthly *dtos.MonthlyTransactionOutput
	err = u.uow.Do(ctx, func(ctx context.Context, tx pkgDatabase.DBExecutor) error {
		// Find or create monthly aggregate
		monthlyAggregate, err := u.repo.FindOrCreateMonthly(ctx, tx, user, referenceMonth)
		if err != nil {
			return fmt.Errorf("failed to find or create monthly transaction: %w", err)
		}

		// Create item using strategy
		item, err := strategy.CreateItem(
			monthlyAggregate.ID,
			categoryID,
			input.Title,
			input.Description,
			amount,
			direction,
			input.IsPaid,
		)
		if err != nil {
			return fmt.Errorf("failed to create transaction item: %w", err)
		}

		// Generate ID for the item
		itemID, _ := sharedVos.NewUUID()
		item.SetID(itemID)

		// Add item to aggregate (recalculates totals)
		if err := monthlyAggregate.AddItem(item); err != nil {
			return fmt.Errorf("failed to add item to aggregate: %w", err)
		}

		// Persist item
		if err := u.repo.InsertItem(ctx, tx, item); err != nil {
			return fmt.Errorf("failed to insert item: %w", err)
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
		u.o11y.Logger().Error(ctx, "failed to register transaction", observability.Error(err))
		return nil, err
	}

	return monthly, nil
}

func (u *registerTransactionUseCase) toOutput(aggregate interface{}) *dtos.MonthlyTransactionOutput {
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
			Amount:      float64(item.Amount.Cents()) / 100.0,
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
		TotalIncome:    float64(monthly.TotalIncome.Cents()) / 100.0,
		TotalExpense:   float64(monthly.TotalExpense.Cents()) / 100.0,
		TotalAmount:    float64(monthly.TotalAmount.Cents()) / 100.0,
		Items:          items,
		CreatedAt:      monthly.CreatedAt.ValueOr(time.Time{}),
		UpdatedAt:      monthly.UpdatedAt.ValueOr(time.Time{}),
	}
}
