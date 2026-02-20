package domain

import (
	"net/http"

	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
)

// ErrorMappings returns the HTTP status mappings for budget domain errors.
// Call NewErrorHandler(o11y, domain.ErrorMappings()) in the budget module.
func ErrorMappings() map[error]httperrors.ErrorMapping {
	return map[error]httperrors.ErrorMapping{
		// Validation errors → 400 Bad Request
		ErrBudgetInvalidTotal: {
			Status:  http.StatusBadRequest,
			Message: "Sum of budget item percentages must equal 100%",
		},
		ErrBudgetPercentageExceeds100: {
			Status:  http.StatusBadRequest,
			Message: "Sum of budget item percentages exceeds 100%",
		},
		ErrBudgetNoItems: {
			Status:  http.StatusBadRequest,
			Message: "Budget must have at least one item",
		},
		ErrInvalidPercentage: {
			Status:  http.StatusBadRequest,
			Message: "Percentage must be between 0 and 100",
		},
		ErrNegativeAmount: {
			Status:  http.StatusBadRequest,
			Message: "Amount cannot be negative",
		},
		ErrInvalidCategoryID: {
			Status:  http.StatusBadRequest,
			Message: "Invalid category ID",
		},

		// Not found errors → 404 Not Found
		ErrBudgetNotFound: {
			Status:  http.StatusNotFound,
			Message: "Budget not found",
		},
		ErrBudgetItemNotFound: {
			Status:  http.StatusNotFound,
			Message: "Budget item not found",
		},

		// Conflict errors → 409 Conflict
		ErrBudgetAlreadyExistsForMonth: {
			Status:  http.StatusConflict,
			Message: "Budget already exists for this month",
		},
		ErrDuplicateCategory: {
			Status:  http.StatusConflict,
			Message: "Category already exists in budget",
		},
	}
}
