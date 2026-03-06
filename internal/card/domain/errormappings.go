package domain

import (
	"net/http"

	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
)

func ErrorMappings() map[error]httperrors.ErrorMapping {
	return map[error]httperrors.ErrorMapping{
		ErrCardHasOpenInvoices:   {Status: http.StatusUnprocessableEntity, Message: "Card has open invoices and cannot be deleted"},
		ErrInvalidCardType:       {Status: http.StatusBadRequest, Message: "Invalid card type"},
		ErrInvalidCardFlag:       {Status: http.StatusBadRequest, Message: "Invalid card flag"},
		ErrInvalidLastFourDigits: {Status: http.StatusBadRequest, Message: "Invalid last four digits"},
		ErrDueDayRequired:        {Status: http.StatusBadRequest, Message: "Due day is required for credit cards"},
	}
}
