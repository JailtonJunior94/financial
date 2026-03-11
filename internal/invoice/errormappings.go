package invoice

import (
	"net/http"

	"github.com/jailtonjunior94/financial/internal/invoice/domain"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
)

// ErrorMappings returns the HTTP status mappings for invoice domain errors.
func ErrorMappings() map[error]httperrors.ErrorMapping {
	return map[error]httperrors.ErrorMapping{
		// Validation errors -> 400 Bad Request
		domain.ErrNegativeAmount: {
			Status:  http.StatusBadRequest,
			Message: "Amount cannot be negative",
		},
		domain.ErrInvalidInstallment: {
			Status:  http.StatusBadRequest,
			Message: "Installment number must be between 1 and installment total",
		},
		domain.ErrInvalidInstallmentTotal: {
			Status:  http.StatusBadRequest,
			Message: "Installment total must be at least 1",
		},
		domain.ErrInstallmentAmountInvalid: {
			Status:  http.StatusBadRequest,
			Message: "Installment amount must equal total amount divided by installments",
		},
		domain.ErrInvalidCategoryID: {
			Status:  http.StatusBadRequest,
			Message: "Invalid category ID",
		},
		domain.ErrInvalidCardID: {
			Status:  http.StatusBadRequest,
			Message: "Invalid card ID",
		},
		domain.ErrEmptyDescription: {
			Status:  http.StatusBadRequest,
			Message: "Description cannot be empty",
		},
		domain.ErrInvoiceHasNoItems: {
			Status:  http.StatusBadRequest,
			Message: "Invoice must have at least one item",
		},
		domain.ErrInvoiceNegativeTotal: {
			Status:  http.StatusBadRequest,
			Message: "Invoice total amount cannot be negative",
		},

		// Ownership errors -> 403 Forbidden
		domain.ErrInvoiceNotOwned: {
			Status:  http.StatusForbidden,
			Message: "Access to this invoice is not allowed",
		},

		// Not found errors -> 404 Not Found
		domain.ErrInvoiceNotFound: {
			Status:  http.StatusNotFound,
			Message: "Invoice not found",
		},
		domain.ErrInvoiceItemNotFound: {
			Status:  http.StatusNotFound,
			Message: "Invoice item not found",
		},

		// Conflict errors -> 409 Conflict
		domain.ErrInvoiceAlreadyExistsForMonth: {
			Status:  http.StatusConflict,
			Message: "Invoice already exists for this card and month",
		},
	}
}
