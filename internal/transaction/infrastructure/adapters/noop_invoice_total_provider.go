package adapters

import (
	"context"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
)

// NoOpInvoiceTotalProvider é um stub que retorna sempre zero.
// Usado para resolver dependência circular durante inicialização.
type NoOpInvoiceTotalProvider struct{}

// NewNoOpInvoiceTotalProvider cria um novo NoOpInvoiceTotalProvider.
func NewNoOpInvoiceTotalProvider() interfaces.InvoiceTotalProvider {
	return &NoOpInvoiceTotalProvider{}
}

// GetClosedInvoiceTotal sempre retorna zero.
func (n *NoOpInvoiceTotalProvider) GetClosedInvoiceTotal(
	ctx context.Context,
	userID sharedVos.UUID,
	referenceMonth transactionVos.ReferenceMonth,
) (sharedVos.Money, error) {
	zero, _ := sharedVos.NewMoney(0, sharedVos.CurrencyBRL)
	return zero, nil
}
