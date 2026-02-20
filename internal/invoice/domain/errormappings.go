package domain

import (
	"net/http"

	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
)

// ErrorMappings returns the HTTP status mappings for invoice domain errors.
// Call NewErrorHandler(o11y, domain.ErrorMappings()) in the invoice module.
func ErrorMappings() map[error]httperrors.ErrorMapping {
	return map[error]httperrors.ErrorMapping{
		// Validation errors → 400 Bad Request
		ErrNegativeAmount: {
			Status:  http.StatusBadRequest,
			Message: "Amount cannot be negative",
		},
		ErrInvalidInstallment: {
			Status:  http.StatusBadRequest,
			Message: "Installment number must be between 1 and installment total",
		},
		ErrInvalidInstallmentTotal: {
			Status:  http.StatusBadRequest,
			Message: "Installment total must be at least 1",
		},
		ErrInstallmentAmountInvalid: {
			Status:  http.StatusBadRequest,
			Message: "Installment amount must equal total amount divided by installments",
		},
		ErrInvalidCategoryID: {
			Status:  http.StatusBadRequest,
			Message: "Invalid category ID",
		},
		ErrInvalidCardID: {
			Status:  http.StatusBadRequest,
			Message: "Invalid card ID",
		},
		ErrEmptyDescription: {
			Status:  http.StatusBadRequest,
			Message: "Description cannot be empty",
		},
		ErrInvoiceHasNoItems: {
			Status:  http.StatusBadRequest,
			Message: "Invoice must have at least one item",
		},
		ErrInvoiceNegativeTotal: {
			Status:  http.StatusBadRequest,
			Message: "Invoice total amount cannot be negative",
		},

		// Not found errors → 404 Not Found
		ErrInvoiceNotFound: {
			Status:  http.StatusNotFound,
			Message: "Invoice not found",
		},
		ErrInvoiceItemNotFound: {
			Status:  http.StatusNotFound,
			Message: "Invoice item not found",
		},

		// Conflict errors → 409 Conflict
		ErrInvoiceAlreadyExistsForMonth: {
			Status:  http.StatusConflict,
			Message: "Invoice already exists for this card and month",
		},
	}
}
