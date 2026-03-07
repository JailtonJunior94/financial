package usecase

import (
	"context"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
	transactionDomain "github.com/jailtonjunior94/financial/internal/transaction/domain"
	transactionInterfaces "github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
)

type (
	GetTransactionUseCase interface {
		Execute(ctx context.Context, userID, transactionID string) (*dtos.TransactionOutput, error)
	}

	getTransactionUseCase struct {
		o11y       observability.Observability
		repository transactionInterfaces.TransactionRepository
	}
)

// NewGetTransactionUseCase creates a new GetTransactionUseCase.
func NewGetTransactionUseCase(
	o11y observability.Observability,
	repository transactionInterfaces.TransactionRepository,
) GetTransactionUseCase {
	return &getTransactionUseCase{o11y: o11y, repository: repository}
}

func (u *getTransactionUseCase) Execute(ctx context.Context, userID, transactionID string) (*dtos.TransactionOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "get_transaction_usecase.execute")
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

	u.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "GetTransaction"),
		observability.String("layer", "usecase"),
		observability.String("entity", "transaction"),
		observability.String("user_id", userID),
	)

	return toOutput(transaction), nil
}
