package domain

import (
	"net/http"

	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
)

// ErrorMappings returns the HTTP status mappings for transaction domain errors.
// Call NewErrorHandler(o11y, domain.ErrorMappings()) in the transaction module.
func ErrorMappings() map[error]httperrors.ErrorMapping {
	return map[error]httperrors.ErrorMapping{
		// Validation errors → 400 Bad Request
		ErrInvalidReferenceMonth: {
			Status:  http.StatusBadRequest,
			Message: "Invalid reference month",
		},
		ErrTransactionItemDeleted: {
			Status:  http.StatusBadRequest,
			Message: "Transaction item has been deleted",
		},
		ErrInvalidTransactionTitle: {
			Status:  http.StatusBadRequest,
			Message: "Invalid transaction title",
		},
		ErrInvalidTransactionAmount: {
			Status:  http.StatusBadRequest,
			Message: "Invalid transaction amount",
		},
		ErrInvalidTransactionDirection: {
			Status:  http.StatusBadRequest,
			Message: "Invalid transaction direction",
		},
		ErrInvalidTransactionType: {
			Status:  http.StatusBadRequest,
			Message: "Invalid transaction type",
		},
		ErrItemDoesNotBelongToMonth: {
			Status:  http.StatusBadRequest,
			Message: "Item does not belong to this monthly transaction",
		},
		ErrCannotUpdateDeletedItem: {
			Status:  http.StatusBadRequest,
			Message: "Cannot update a deleted item",
		},
		ErrCannotDeleteDeletedItem: {
			Status:  http.StatusBadRequest,
			Message: "Cannot delete an already deleted item",
		},
		ErrNegativeAmount: {
			Status:  http.StatusBadRequest,
			Message: "Amount cannot be negative",
		},

		// Not found errors → 404 Not Found
		ErrMonthlyTransactionNotFound: {
			Status:  http.StatusNotFound,
			Message: "Monthly transaction not found",
		},
		ErrTransactionItemNotFound: {
			Status:  http.StatusNotFound,
			Message: "Transaction item not found",
		},

		// Conflict errors → 409 Conflict
		ErrMonthlyTransactionAlreadyExists: {
			Status:  http.StatusConflict,
			Message: "Monthly transaction already exists for this month",
		},

		// Service errors → 503 Service Unavailable
		ErrInvoiceProviderUnavailable: {
			Status:  http.StatusServiceUnavailable,
			Message: "Invoice provider is currently unavailable",
		},
	}
}
