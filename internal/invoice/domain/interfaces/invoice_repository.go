package interfaces

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/invoice/domain/entities"
	invoiceVos "github.com/jailtonjunior94/financial/internal/invoice/domain/vos"
)

// InvoiceRepository define as operações de persistência de Invoice
type InvoiceRepository interface {
	// Insert cria uma nova fatura
	Insert(ctx context.Context, invoice *entities.Invoice) error

	// InsertItems cria múltiplos itens de fatura
	InsertItems(ctx context.Context, items []*entities.InvoiceItem) error

	// FindByID busca uma fatura por ID
	FindByID(ctx context.Context, id vos.UUID) (*entities.Invoice, error)

	// FindByUserAndCardAndMonth busca fatura por usuário, cartão e mês de referência
	FindByUserAndCardAndMonth(
		ctx context.Context,
		userID vos.UUID,
		cardID vos.UUID,
		referenceMonth invoiceVos.ReferenceMonth,
	) (*entities.Invoice, error)

	// FindByUserAndMonth busca todas as faturas de um usuário em um mês
	FindByUserAndMonth(
		ctx context.Context,
		userID vos.UUID,
		referenceMonth invoiceVos.ReferenceMonth,
	) ([]*entities.Invoice, error)

	// FindByCard busca todas as faturas de um cartão
	FindByCard(ctx context.Context, cardID vos.UUID) ([]*entities.Invoice, error)

	// Update atualiza uma fatura
	Update(ctx context.Context, invoice *entities.Invoice) error

	// UpdateItem atualiza um item de fatura
	UpdateItem(ctx context.Context, item *entities.InvoiceItem) error

	// DeleteItem remove um item (soft delete)
	DeleteItem(ctx context.Context, itemID vos.UUID) error

	// FindItemsByPurchaseOrigin busca todos os itens relacionados a uma compra original
	// (útil para atualizar/deletar compras parceladas)
	FindItemsByPurchaseOrigin(
		ctx context.Context,
		purchaseDate string,
		categoryID vos.UUID,
		description string,
	) ([]*entities.InvoiceItem, error)
}
