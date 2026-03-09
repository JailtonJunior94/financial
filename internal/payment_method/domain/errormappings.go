package domain

import (
	"net/http"

	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
)

// ErrorMappings returns the HTTP status mappings for payment_method domain errors.
// Call NewErrorHandler(o11y, domain.ErrorMappings()) in the payment_method module.
func ErrorMappings() map[error]httperrors.ErrorMapping {
	return map[error]httperrors.ErrorMapping{
		ErrPaymentMethodNotFound: {
			Status:  http.StatusNotFound,
			Message: "Payment method not found",
		},
	}
}
