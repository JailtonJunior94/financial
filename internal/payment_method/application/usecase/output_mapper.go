package usecase

import (
	"time"

	"github.com/jailtonjunior94/financial/internal/payment_method/application/dtos"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/entities"
)

func toPaymentMethodOutput(paymentMethod *entities.PaymentMethod) *dtos.PaymentMethodOutput {
	output := &dtos.PaymentMethodOutput{
		ID:          paymentMethod.ID.String(),
		Name:        paymentMethod.Name.String(),
		Code:        paymentMethod.Code.String(),
		Description: paymentMethod.Description.String(),
		CreatedAt:   paymentMethod.CreatedAt.ValueOr(time.Time{}),
	}

	if !paymentMethod.UpdatedAt.ValueOr(time.Time{}).IsZero() {
		output.UpdatedAt = paymentMethod.UpdatedAt.ValueOr(time.Time{})
	}

	return output
}
