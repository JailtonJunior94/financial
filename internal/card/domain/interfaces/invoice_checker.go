package interfaces

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type InvoiceChecker interface {
	HasOpenInvoices(ctx context.Context, cardID vos.UUID) (bool, error)
}
