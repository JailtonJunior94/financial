package interfaces

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/payment_method/domain/entities"
	"github.com/jailtonjunior94/financial/pkg/pagination"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

// ListPaymentMethodsParams representa os parâmetros para paginação de payment methods.
type ListPaymentMethodsParams struct {
	Limit  int
	Cursor pagination.Cursor
	Code   string // Opcional: filtra por código
}

type PaymentMethodRepository interface {
	List(ctx context.Context) ([]*entities.PaymentMethod, error)
	ListPaginated(ctx context.Context, params ListPaymentMethodsParams) ([]*entities.PaymentMethod, error)
	FindByID(ctx context.Context, id vos.UUID) (*entities.PaymentMethod, error)
	FindByCode(ctx context.Context, code string) (*entities.PaymentMethod, error)
	Save(ctx context.Context, paymentMethod *entities.PaymentMethod) error
	Update(ctx context.Context, paymentMethod *entities.PaymentMethod) error
}
