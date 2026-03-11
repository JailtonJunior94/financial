package card

import (
	"net/http"

	"github.com/jailtonjunior94/financial/internal/card/domain"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
)

// ErrorMappings returns the HTTP status mappings for card domain errors.
func ErrorMappings() map[error]httperrors.ErrorMapping {
	return map[error]httperrors.ErrorMapping{
		domain.ErrCardNotFound:          {Status: http.StatusNotFound, Message: "Card not found"},
		domain.ErrCardHasOpenInvoices:   {Status: http.StatusUnprocessableEntity, Message: "Card has open invoices and cannot be deleted"},
		domain.ErrInvalidCardType:       {Status: http.StatusBadRequest, Message: "Invalid card type"},
		domain.ErrInvalidCardFlag:       {Status: http.StatusBadRequest, Message: "Invalid card flag"},
		domain.ErrInvalidLastFourDigits: {Status: http.StatusBadRequest, Message: "Invalid last four digits"},
		domain.ErrDueDayRequired:        {Status: http.StatusBadRequest, Message: "Due day is required for credit cards"},
	}
}
