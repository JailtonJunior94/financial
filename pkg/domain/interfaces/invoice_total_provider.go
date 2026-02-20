package interfaces

import (
	"context"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

// InvoiceTotalProvider fornece informações sobre faturas fechadas.
// Interface compartilhada entre os módulos de invoice e transaction (Port & Adapter).
type InvoiceTotalProvider interface {
	// GetClosedInvoiceTotal retorna o total da fatura fechada para o mês.
	// Se não houver fatura ou ela não estiver fechada, retorna zero.
	GetClosedInvoiceTotal(
		ctx context.Context,
		userID sharedVos.UUID,
		referenceMonth pkgVos.ReferenceMonth,
	) (sharedVos.Money, error)
}
