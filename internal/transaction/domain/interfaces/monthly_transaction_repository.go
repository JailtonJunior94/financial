package interfaces

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

// MonthlyTransactionRepository define as operações de persistência para transações mensais.
type MonthlyTransactionRepository interface {
	// FindByUserAndMonth busca uma transação mensal por usuário e mês de referência
	FindByUserAndMonth(
		ctx context.Context,
		executor database.DBTX,
		userID vos.UUID,
		referenceMonth pkgVos.ReferenceMonth,
	) (*entities.MonthlyTransaction, error)

	// Save persiste uma nova transação mensal
	Save(
		ctx context.Context,
		executor database.DBTX,
		monthlyTransaction *entities.MonthlyTransaction,
	) error

	// Update atualiza uma transação mensal existente
	Update(
		ctx context.Context,
		executor database.DBTX,
		monthlyTransaction *entities.MonthlyTransaction,
	) error

	// SaveItem persiste um novo item de transação
	SaveItem(
		ctx context.Context,
		executor database.DBTX,
		item *entities.TransactionItem,
	) error

	// UpdateItem atualiza um item de transação existente
	UpdateItem(
		ctx context.Context,
		executor database.DBTX,
		item *entities.TransactionItem,
	) error

	// FindItemByID busca um item por ID
	FindItemByID(
		ctx context.Context,
		executor database.DBTX,
		itemID vos.UUID,
	) (*entities.TransactionItem, error)
}
