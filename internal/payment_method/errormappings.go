package payment_method

import (
	"net/http"

	"github.com/jailtonjunior94/financial/internal/payment_method/domain"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
)

// ErrorMappings returns the HTTP status mappings for payment_method domain errors.
func ErrorMappings() map[error]httperrors.ErrorMapping {
	return map[error]httperrors.ErrorMapping{
		domain.ErrPaymentMethodNotFound: {
			Status:  http.StatusNotFound,
			Message: "Payment method not found",
		},
	}
}
