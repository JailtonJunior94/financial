package interfaces

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/database"
)

// MonthlyTransactionRepository define as operações de persistência para transações mensais
type MonthlyTransactionRepository interface {
	// FindByUserAndMonth busca uma transação mensal por usuário e mês de referência
	FindByUserAndMonth(
		ctx context.Context,
		executor database.DBExecutor,
		userID vos.UUID,
		referenceMonth transactionVos.ReferenceMonth,
	) (*entities.MonthlyTransaction, error)

	// Save persiste uma nova transação mensal
	Save(
		ctx context.Context,
		executor database.DBExecutor,
		monthlyTransaction *entities.MonthlyTransaction,
	) error

	// Update atualiza uma transação mensal existente
	Update(
		ctx context.Context,
		executor database.DBExecutor,
		monthlyTransaction *entities.MonthlyTransaction,
	) error

	// SaveItem persiste um novo item de transação
	SaveItem(
		ctx context.Context,
		executor database.DBExecutor,
		item *entities.TransactionItem,
	) error

	// UpdateItem atualiza um item de transação existente
	UpdateItem(
		ctx context.Context,
		executor database.DBExecutor,
		item *entities.TransactionItem,
	) error

	// FindItemByID busca um item por ID
	FindItemByID(
		ctx context.Context,
		executor database.DBExecutor,
		itemID vos.UUID,
	) (*entities.TransactionItem, error)
}
