package interfaces

import (
	"context"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

// InvoiceInfo holds the invoice data needed by the transaction module.
type InvoiceInfo struct {
	ID     vos.UUID
	Status string
}

// InvoiceProvider defines the contract for invoice operations consumed by the transaction module.
type InvoiceProvider interface {
	FindOrCreate(ctx context.Context, userID, cardID vos.UUID, referenceMonth pkgVos.ReferenceMonth, dueDate time.Time) (*InvoiceInfo, error)
	GetStatus(ctx context.Context, invoiceID vos.UUID) (string, error)
}
