package httperrors

import (
	"errors"
	"net/http"
	"testing"

	budgetdomain "github.com/jailtonjunior94/financial/internal/budget/domain"
	invoicedomain "github.com/jailtonjunior94/financial/internal/invoice/domain"
	transactiondomain "github.com/jailtonjunior94/financial/internal/transaction/domain"
	customerrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
)

func TestErrorMapper_MapError_ValidationErrors(t *testing.T) {
	mapper := NewErrorMapper()

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		description    string
	}{
		// Budget validation errors should return 400
		{
			name:           "ErrBudgetInvalidTotal returns 400",
			err:            budgetdomain.ErrBudgetInvalidTotal,
			expectedStatus: http.StatusBadRequest,
			description:    "Budget validation errors should return 400 Bad Request, not 500",
		},
		{
			name:           "ErrBudgetPercentageExceeds100 returns 400",
			err:            budgetdomain.ErrBudgetPercentageExceeds100,
			expectedStatus: http.StatusBadRequest,
			description:    "Budget percentage validation should return 400 Bad Request",
		},
		{
			name:           "ErrBudgetNoItems returns 400",
			err:            budgetdomain.ErrBudgetNoItems,
			expectedStatus: http.StatusBadRequest,
			description:    "Budget items validation should return 400 Bad Request",
		},
		{
			name:           "ErrInvalidPercentage returns 400",
			err:            budgetdomain.ErrInvalidPercentage,
			expectedStatus: http.StatusBadRequest,
			description:    "Invalid percentage should return 400 Bad Request",
		},
		{
			name:           "ErrDuplicateCategory returns 409",
			err:            budgetdomain.ErrDuplicateCategory,
			expectedStatus: http.StatusConflict,
			description:    "Duplicate category should return 409 Conflict",
		},

		// Budget not found errors should return 404
		{
			name:           "ErrBudgetNotFound returns 404",
			err:            budgetdomain.ErrBudgetNotFound,
			expectedStatus: http.StatusNotFound,
			description:    "Budget not found should return 404 Not Found",
		},
		{
			name:           "ErrBudgetItemNotFound returns 404",
			err:            budgetdomain.ErrBudgetItemNotFound,
			expectedStatus: http.StatusNotFound,
			description:    "Budget item not found should return 404 Not Found",
		},

		// Transaction validation errors should return 400
		{
			name:           "ErrInvalidReferenceMonth returns 400",
			err:            transactiondomain.ErrInvalidReferenceMonth,
			expectedStatus: http.StatusBadRequest,
			description:    "Invalid reference month should return 400 Bad Request",
		},
		{
			name:           "ErrCannotUpdateDeletedItem returns 400",
			err:            transactiondomain.ErrCannotUpdateDeletedItem,
			expectedStatus: http.StatusBadRequest,
			description:    "Cannot update deleted item should return 400 Bad Request",
		},
		{
			name:           "ErrTransactionItemNotFound returns 404",
			err:            transactiondomain.ErrTransactionItemNotFound,
			expectedStatus: http.StatusNotFound,
			description:    "Transaction item not found should return 404 Not Found",
		},

		// Invoice validation errors should return 400
		{
			name:           "ErrInvalidPurchaseDate returns 400",
			err:            invoicedomain.ErrInvalidPurchaseDate,
			expectedStatus: http.StatusBadRequest,
			description:    "Invalid purchase date should return 400 Bad Request",
		},
		{
			name:           "ErrInvoiceHasNoItems returns 400",
			err:            invoicedomain.ErrInvoiceHasNoItems,
			expectedStatus: http.StatusBadRequest,
			description:    "Invoice with no items should return 400 Bad Request",
		},

		// Service errors should return 503
		{
			name:           "ErrInvoiceProviderUnavailable returns 503",
			err:            transactiondomain.ErrInvoiceProviderUnavailable,
			expectedStatus: http.StatusServiceUnavailable,
			description:    "Service unavailable should return 503 Service Unavailable",
		},

		// Custom errors validation
		{
			name:           "customerrors.ErrBudgetInvalidTotal returns 400",
			err:            customerrors.ErrBudgetInvalidTotal,
			expectedStatus: http.StatusBadRequest,
			description:    "Custom error budget validation should return 400 Bad Request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := mapper.MapError(tt.err)

			if mapping.Status != tt.expectedStatus {
				t.Errorf("%s: expected status %d, got %d. Error: %v",
					tt.description,
					tt.expectedStatus,
					mapping.Status,
					tt.err,
				)
			}
		})
	}
}

func TestErrorMapper_MapError_UnmappedErrors(t *testing.T) {
	mapper := NewErrorMapper()

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		description    string
	}{
		{
			name:           "Unmapped error returns 500",
			err:            errors.New("some unknown error"),
			expectedStatus: http.StatusInternalServerError,
			description:    "Unknown errors should return 500 Internal Server Error",
		},
		{
			name:           "Nil error returns 500",
			err:            nil,
			expectedStatus: http.StatusInternalServerError,
			description:    "Nil errors should return 500 Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := mapper.MapError(tt.err)

			if mapping.Status != tt.expectedStatus {
				t.Errorf("%s: expected status %d, got %d",
					tt.description,
					tt.expectedStatus,
					mapping.Status,
				)
			}
		})
	}
}
