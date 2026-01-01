package interfaces

import (
	"context"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
)

// InvoiceTotalProvider fornece informações sobre faturas fechadas.
// Interface de domínio para integração com o módulo de invoices (Port & Adapter).
type InvoiceTotalProvider interface {
	// GetClosedInvoiceTotal retorna o total da fatura fechada para o mês.
	// Se não houver fatura ou ela não estiver fechada, retorna zero.
	GetClosedInvoiceTotal(
		ctx context.Context,
		userID sharedVos.UUID,
		referenceMonth transactionVos.ReferenceMonth,
	) (sharedVos.Money, error)
}
