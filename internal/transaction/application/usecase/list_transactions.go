package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
	transactionInterfaces "github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type (
	ListTransactionsUseCase interface {
		Execute(ctx context.Context, userID string, params *dtos.ListParams) (*dtos.TransactionListOutput, error)
	}

	listTransactionsUseCase struct {
		o11y       observability.Observability
		repository transactionInterfaces.TransactionRepository
	}
)

// NewListTransactionsUseCase creates a new ListTransactionsUseCase.
func NewListTransactionsUseCase(
	o11y observability.Observability,
	repository transactionInterfaces.TransactionRepository,
) ListTransactionsUseCase {
	return &listTransactionsUseCase{o11y: o11y, repository: repository}
}

func (u *listTransactionsUseCase) Execute(ctx context.Context, userID string, params *dtos.ListParams) (*dtos.TransactionListOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "list_transactions_usecase.execute")
	defer span.End()

	userUUID, err := vos.NewUUIDFromString(userID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	limit := params.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	repoParams := transactionInterfaces.ListParams{
		UserID:        userUUID,
		PaymentMethod: params.PaymentMethod,
		CategoryID:    params.CategoryID,
		Limit:         limit,
		Cursor:        params.Cursor,
	}

	if params.StartDate != "" {
		t, err := time.Parse("2006-01-02", params.StartDate)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("invalid start_date: %w", err)
		}
		repoParams.StartDate = &t
	}
	if params.EndDate != "" {
		t, err := time.Parse("2006-01-02", params.EndDate)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("invalid end_date: %w", err)
		}
		repoParams.EndDate = &t
	}

	transactions, nextCursor, err := u.repository.ListPaginated(ctx, repoParams)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	u.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "ListTransactions"),
		observability.String("layer", "usecase"),
		observability.String("entity", "transaction"),
		observability.String("user_id", userID),
	)

	return &dtos.TransactionListOutput{
		Data:       toOutputList(transactions),
		NextCursor: nextCursor,
	}, nil
}
