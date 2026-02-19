package usecase

import (
	"context"
	"fmt"
	"time"

	appstrategies "github.com/jailtonjunior94/financial/internal/transaction/application/strategies"
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
		ccItemPersister      appstrategies.CreditCardItemPersister
		o11y                 observability.Observability
	}
)

func NewRegisterTransactionUseCase(
	uow uow.UnitOfWork,
	repo interfaces.TransactionRepository,
	invoiceTotalProvider interfaces.InvoiceTotalProvider,
	ccItemPersister appstrategies.CreditCardItemPersister,
	o11y observability.Observability,
) RegisterTransactionUseCase {
	return &registerTransactionUseCase{
		uow:                  uow,
		repo:                 repo,
		invoiceTotalProvider: invoiceTotalProvider,
		ccItemPersister:      ccItemPersister,
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

	user, err := sharedVos.NewUUIDFromString(userID)
	if err != nil {
		u.o11y.Logger().Error(ctx, "invalid user id", observability.Error(err))
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	categoryID, err := sharedVos.NewUUIDFromString(input.CategoryID)
	if err != nil {
		u.o11y.Logger().Error(ctx, "invalid category id", observability.Error(err))
		return nil, fmt.Errorf("invalid category ID: %w", err)
	}

	referenceMonth, err := transactionVos.NewReferenceMonthFromString(input.ReferenceMonth)
	if err != nil {
		u.o11y.Logger().Error(ctx, "invalid reference month", observability.Error(err))
		return nil, fmt.Errorf("invalid reference month: %w", err)
	}

	direction, err := transactionVos.NewTransactionDirection(input.Direction)
	if err != nil {
		u.o11y.Logger().Error(ctx, "invalid direction", observability.Error(err))
		return nil, fmt.Errorf("invalid direction: %w", err)
	}

	transactionType, err := transactionVos.NewTransactionType(input.Type)
	if err != nil {
		u.o11y.Logger().Error(ctx, "invalid transaction type", observability.Error(err))
		return nil, fmt.Errorf("invalid transaction type: %w", err)
	}

	amount, err := sharedVos.NewMoneyFromString(input.Amount, sharedVos.CurrencyBRL)
	if err != nil {
		u.o11y.Logger().Error(ctx, "invalid amount", observability.Error(err))
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	var monthly *dtos.MonthlyTransactionOutput
	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		monthlyAggregate, err := u.repo.FindOrCreateMonthly(ctx, tx, user, referenceMonth)
		if err != nil {
			return fmt.Errorf("failed to find or create monthly transaction: %w", err)
		}

		if transactionType == transactionVos.TypeCreditCard {
			invoiceTotal, err := u.invoiceTotalProvider.GetClosedInvoiceTotal(ctx, user, referenceMonth)
			if err != nil {
				return fmt.Errorf("failed to get closed invoice total: %w", err)
			}

			u.o11y.Logger().Info(ctx, "fetched closed invoice total",
				observability.String("user_id", user.String()),
				observability.String("reference_month", referenceMonth.String()),
				observability.Int64("total_cents", invoiceTotal.Cents()),
			)

			// Snapshot de IDs existentes antes da mutação do aggregate (determina INSERT vs UPDATE)
			existingIDs := make(map[string]struct{}, len(monthlyAggregate.Items))
			for _, item := range monthlyAggregate.Items {
				existingIDs[item.ID.String()] = struct{}{}
			}

			if err := monthlyAggregate.UpdateOrCreateCreditCardItem(categoryID, invoiceTotal, input.IsPaid); err != nil {
				return fmt.Errorf("failed to update or create credit card item: %w", err)
			}

			var creditCardItem *entities.TransactionItem
			for _, item := range monthlyAggregate.Items {
				if !item.Type.IsCreditCard() || item.IsDeleted() {
					continue
				}
				creditCardItem = item
				break
			}

			if creditCardItem == nil {
				return fmt.Errorf("credit card item not found after update/create")
			}

			if err := u.ccItemPersister.Persist(ctx, tx, monthlyAggregate, existingIDs); err != nil {
				return fmt.Errorf("failed to persist credit card items: %w", err)
			}

			if err := u.repo.UpdateMonthly(ctx, tx, monthlyAggregate); err != nil {
				return fmt.Errorf("failed to update monthly transaction: %w", err)
			}

			monthly = u.toOutput(monthlyAggregate)
			return nil
		}

		strategy := strategies.GetStrategy(transactionType)
		if strategy == nil {
			u.o11y.Logger().Error(ctx, "strategy not found", observability.String("type", input.Type))
			return fmt.Errorf("unsupported transaction type: %s", input.Type)
		}

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

		itemID, _ := sharedVos.NewUUID()
		item.SetID(itemID)

		if err := monthlyAggregate.AddItem(item); err != nil {
			return fmt.Errorf("failed to add item to aggregate: %w", err)
		}

		if err := u.repo.InsertItem(ctx, tx, item); err != nil {
			return fmt.Errorf("failed to insert item: %w", err)
		}

		if err := u.repo.UpdateMonthly(ctx, tx, monthlyAggregate); err != nil {
			return fmt.Errorf("failed to update monthly totals: %w", err)
		}

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
