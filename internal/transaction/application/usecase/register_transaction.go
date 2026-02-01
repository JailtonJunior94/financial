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
	RegisterTransactionUseCase interface {
		Execute(ctx context.Context, userID string, input *dtos.RegisterTransactionInput) (*dtos.MonthlyTransactionOutput, error)
	}

	registerTransactionUseCase struct {
		uow                  uow.UnitOfWork
		repo                 interfaces.TransactionRepository
		invoiceTotalProvider interfaces.InvoiceTotalProvider
		o11y                 observability.Observability
	}
)

func NewRegisterTransactionUseCase(
	uow uow.UnitOfWork,
	repo interfaces.TransactionRepository,
	invoiceTotalProvider interfaces.InvoiceTotalProvider,
	o11y observability.Observability,
) RegisterTransactionUseCase {
	return &registerTransactionUseCase{
		uow:                  uow,
		repo:                 repo,
		invoiceTotalProvider: invoiceTotalProvider,
		o11y:                 o11y,
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
	amount, err := sharedVos.NewMoneyFromString(input.Amount, sharedVos.CurrencyBRL)
	if err != nil {
		u.o11y.Logger().Error(ctx, "invalid amount", observability.Error(err))
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	// Execute within transaction
	var monthly *dtos.MonthlyTransactionOutput
	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		// Find or create monthly aggregate
		monthlyAggregate, err := u.repo.FindOrCreateMonthly(ctx, tx, user, referenceMonth)
		if err != nil {
			return fmt.Errorf("failed to find or create monthly transaction: %w", err)
		}

		// Special handling for CREDIT_CARD transactions
		if transactionType == transactionVos.TypeCreditCard {
			// Get invoice total from invoice module
			invoiceTotal, err := u.invoiceTotalProvider.GetClosedInvoiceTotal(ctx, user, referenceMonth)
			if err != nil {
				return fmt.Errorf("failed to get closed invoice total: %w", err)
			}

			u.o11y.Logger().Info(ctx, "fetched closed invoice total",
				observability.String("user_id", user.String()),
				observability.String("reference_month", referenceMonth.String()),
				observability.Int64("total_cents", invoiceTotal.Cents()),
			)

			// Update or create credit card item with invoice total
			if err := monthlyAggregate.UpdateOrCreateCreditCardItem(categoryID, invoiceTotal, input.IsPaid); err != nil {
				return fmt.Errorf("failed to update or create credit card item: %w", err)
			}

			// Find the credit card item to publish event
			// Loop through items to find the CREDIT_CARD type
			var creditCardItem *entities.TransactionItem
			for _, item := range monthlyAggregate.Items {
				if item.Type == transactionVos.TypeCreditCard && !item.IsDeleted() {
					creditCardItem = item
					break
				}
			}

			if creditCardItem == nil {
				return fmt.Errorf("credit card item not found after update/create")
			}

			// Persist changes (update or insert)
			// Note: UpdateOrCreateCreditCardItem handles both cases internally
			if err := u.repo.UpdateMonthly(ctx, tx, monthlyAggregate); err != nil {
				return fmt.Errorf("failed to update monthly transaction: %w", err)
			}

			// Convert to DTO
			monthly = u.toOutput(monthlyAggregate)
			return nil
		}

		// Standard flow for non-credit-card transactions
		// Get strategy for validation and creation
		strategy := strategies.GetStrategy(transactionType)
		if strategy == nil {
			u.o11y.Logger().Error(ctx, "strategy not found", observability.String("type", input.Type))
			return fmt.Errorf("unsupported transaction type: %s", input.Type)
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

func (u *registerTransactionUseCase) toOutput(aggregate any) *dtos.MonthlyTransactionOutput {
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
