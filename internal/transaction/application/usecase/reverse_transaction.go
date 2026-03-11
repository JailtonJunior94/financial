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
	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	transactionInterfaces "github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
)

type (
	ReverseTransactionUseCase interface {
		Execute(ctx context.Context, userID, transactionID string) (*dtos.ReverseOutput, error)
	}

	reverseTransactionUseCase struct {
		o11y            observability.Observability
		uow             uow.UnitOfWork
		repository      transactionInterfaces.TransactionRepository
		invoiceProvider transactionInterfaces.InvoiceProvider
	}
)

// NewReverseTransactionUseCase creates a new ReverseTransactionUseCase.
func NewReverseTransactionUseCase(
	o11y observability.Observability,
	unitOfWork uow.UnitOfWork,
	repository transactionInterfaces.TransactionRepository,
	invoiceProvider transactionInterfaces.InvoiceProvider,
) ReverseTransactionUseCase {
	return &reverseTransactionUseCase{
		o11y:            o11y,
		uow:             unitOfWork,
		repository:      repository,
		invoiceProvider: invoiceProvider,
	}
}

func (u *reverseTransactionUseCase) Execute(ctx context.Context, userID, transactionID string) (*dtos.ReverseOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "reverse_transaction_usecase.execute")
	defer span.End()

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

	scope := []*entities.Transaction{transaction}
	if transaction.InstallmentGroupID != nil {
		scope, err = u.repository.FindByInstallmentGroup(ctx, *transaction.InstallmentGroupID)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
	}

	cancelled := make([]*entities.Transaction, 0)
	kept := make([]*entities.Transaction, 0)

	for _, t := range scope {
		if t.InvoiceID == nil {
			if err := t.Cancel(); err != nil {
				span.RecordError(err)
				return nil, err
			}
			cancelled = append(cancelled, t)
			continue
		}
		status, err := u.invoiceProvider.GetStatus(ctx, *t.InvoiceID)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		if t.IsEditable(status) {
			if err := t.Cancel(); err != nil {
				span.RecordError(err)
				return nil, err
			}
			cancelled = append(cancelled, t)
		} else {
			kept = append(kept, t)
		}
	}

	if len(cancelled) == 0 {
		return nil, transactionDomain.ErrNothingToReverse
	}

	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		return u.repository.UpdateAll(ctx, tx, cancelled)
	})
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	u.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "ReverseTransaction"),
		observability.String("layer", "usecase"),
		observability.String("entity", "transaction"),
		observability.String("user_id", userID),
	)

	return &dtos.ReverseOutput{
		Cancelled: toOutputList(cancelled),
		Kept:      toOutputList(kept),
	}, nil
}
