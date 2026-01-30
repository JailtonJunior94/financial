package interfaces

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/pagination"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
)

// ListMonthlyParams representa os parâmetros para paginação de monthly transactions.
type ListMonthlyParams struct {
	UserID sharedVos.UUID
	Limit  int
	Cursor pagination.Cursor
}

// TransactionRepository define o contrato de persistência para transações.
type TransactionRepository interface {
	// FindOrCreateMonthly busca ou cria o aggregate do mês.
	FindOrCreateMonthly(
		ctx context.Context,
		executor database.DBTX,
		userID sharedVos.UUID,
		referenceMonth transactionVos.ReferenceMonth,
	) (*entities.MonthlyTransaction, error)

	// FindMonthlyByID busca o aggregate por ID com todos os items.
	FindMonthlyByID(
		ctx context.Context,
		executor database.DBTX,
		userID sharedVos.UUID,
		monthlyID sharedVos.UUID,
	) (*entities.MonthlyTransaction, error)

	// UpdateMonthly atualiza o aggregate (totais).
	UpdateMonthly(
		ctx context.Context,
		executor database.DBTX,
		monthly *entities.MonthlyTransaction,
	) error

	// InsertItem insere um novo transaction item.
	InsertItem(
		ctx context.Context,
		executor database.DBTX,
		item *entities.TransactionItem,
	) error

	// UpdateItem atualiza um transaction item existente.
	UpdateItem(
		ctx context.Context,
		executor database.DBTX,
		item *entities.TransactionItem,
	) error

	// FindItemByID busca um item por ID.
	FindItemByID(
		ctx context.Context,
		executor database.DBTX,
		userID sharedVos.UUID,
		itemID sharedVos.UUID,
	) (*entities.TransactionItem, error)

	// ListMonthlyPaginated lista monthly transactions com paginação cursor-based.
	ListMonthlyPaginated(ctx context.Context, params ListMonthlyParams) ([]*entities.MonthlyTransaction, error)

	// GetMonthlyByID busca um monthly transaction por ID (sem executor UoW).
	GetMonthlyByID(ctx context.Context, userID sharedVos.UUID, monthlyID sharedVos.UUID) (*entities.MonthlyTransaction, error)
}
