package budget

import (
	"net/http"

	"github.com/jailtonjunior94/financial/internal/budget/domain"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
)

// ErrorMappings returns the HTTP status mappings for budget domain errors.
func ErrorMappings() map[error]httperrors.ErrorMapping {
	return map[error]httperrors.ErrorMapping{
		// Category validation errors -> 400 Bad Request
		domain.ErrCategoryNotFound: {
			Status:  http.StatusBadRequest,
			Message: "One or more categories not found",
		},
		domain.ErrCategoryNotOwnedByUser: {
			Status:  http.StatusBadRequest,
			Message: "One or more categories do not belong to user",
		},

		// Validation errors -> 400 Bad Request
		domain.ErrBudgetInvalidTotal: {
			Status:  http.StatusBadRequest,
			Message: "Sum of budget item percentages must equal 100%",
		},
		domain.ErrBudgetPercentageExceeds100: {
			Status:  http.StatusBadRequest,
			Message: "Sum of budget item percentages exceeds 100%",
		},
		domain.ErrBudgetNoItems: {
			Status:  http.StatusBadRequest,
			Message: "Budget must have at least one item",
		},
		domain.ErrInvalidPercentage: {
			Status:  http.StatusBadRequest,
			Message: "Percentage must be between 0 and 100",
		},
		domain.ErrNegativeAmount: {
			Status:  http.StatusBadRequest,
			Message: "Amount cannot be negative",
		},
		domain.ErrInvalidCategoryID: {
			Status:  http.StatusBadRequest,
			Message: "Invalid category ID",
		},

		// Not found errors -> 404 Not Found
		domain.ErrBudgetNotFound: {
			Status:  http.StatusNotFound,
			Message: "Budget not found",
		},
		domain.ErrBudgetItemNotFound: {
			Status:  http.StatusNotFound,
			Message: "Budget item not found",
		},

		// Conflict errors -> 409 Conflict
		domain.ErrBudgetAlreadyExistsForMonth: {
			Status:  http.StatusConflict,
			Message: "Budget already exists for this month",
		},
		domain.ErrDuplicateCategory: {
			Status:  http.StatusConflict,
			Message: "Category already exists in budget",
		},
	}
}
