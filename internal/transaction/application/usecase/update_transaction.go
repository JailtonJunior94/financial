package usecase

import (
	"context"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
	transactionDomain "github.com/jailtonjunior94/financial/internal/transaction/domain"
	transactionInterfaces "github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
)

type (
	UpdateTransactionUseCase interface {
		Execute(ctx context.Context, userID, transactionID string, input *dtos.TransactionUpdateInput) (*dtos.TransactionOutput, error)
	}

	updateTransactionUseCase struct {
		o11y            observability.Observability
		uow             uow.UnitOfWork
		repository      transactionInterfaces.TransactionRepository
		invoiceProvider transactionInterfaces.InvoiceProvider
	}
)

// NewUpdateTransactionUseCase creates a new UpdateTransactionUseCase.
func NewUpdateTransactionUseCase(
	o11y observability.Observability,
	unitOfWork uow.UnitOfWork,
	repository transactionInterfaces.TransactionRepository,
	invoiceProvider transactionInterfaces.InvoiceProvider,
) UpdateTransactionUseCase {
	return &updateTransactionUseCase{
		o11y:            o11y,
		uow:             unitOfWork,
		repository:      repository,
		invoiceProvider: invoiceProvider,
	}
}

func (u *updateTransactionUseCase) Execute(ctx context.Context, userID, transactionID string, input *dtos.TransactionUpdateInput) (*dtos.TransactionOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "update_transaction_usecase.execute")
	defer span.End()

	if err := input.Validate(); err != nil {
		span.RecordError(err)
		return nil, err
	}

	txID, err := vos.NewUUIDFromString(transactionID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("invalid transaction_id: %w", err)
	}

	transaction, err := u.repository.FindByID(ctx, txID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	if transaction == nil {
		return nil, transactionDomain.ErrTransactionNotFound
	}

	if transaction.UserID.String() != userID {
		return nil, transactionDomain.ErrTransactionNotOwned
	}

	if transaction.InvoiceID != nil {
		status, err := u.invoiceProvider.GetStatus(ctx, *transaction.InvoiceID)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		if !transaction.IsEditable(status) {
			return nil, transactionDomain.ErrInvoiceClosed
		}
	}

	categoryID, err := vos.NewUUIDFromString(input.CategoryID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("invalid category_id: %w", err)
	}

	amount, err := vos.NewMoneyFromFloat(input.Amount, vos.CurrencyBRL)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	if err := transaction.UpdateDetails(input.Description, amount, categoryID); err != nil {
		span.RecordError(err)
		return nil, err
	}

	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		return u.repository.Update(ctx, tx, transaction)
	})
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	u.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "UpdateTransaction"),
		observability.String("layer", "usecase"),
		observability.String("entity", "transaction"),
		observability.String("user_id", userID),
	)

	return toOutput(transaction), nil
}
