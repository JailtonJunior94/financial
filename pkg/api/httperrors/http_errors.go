package httperrors

import (
	"errors"
	"net/http"

	financialErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
)

var responseErrors = map[error]*ResponseError{
	// Authentication errors
	financialErrors.ErrUnauthorized: {
		Code:    http.StatusUnauthorized,
		Message: "Unauthorized",
	},
	financialErrors.ErrMissingAuthHeader: {
		Code:    http.StatusUnauthorized,
		Message: "Unauthorized",
	},
	financialErrors.ErrInvalidAuthFormat: {
		Code:    http.StatusUnauthorized,
		Message: "Unauthorized",
	},
	financialErrors.ErrInvalidToken: {
		Code:    http.StatusUnauthorized,
		Message: "Unauthorized",
	},
	financialErrors.ErrTokenExpired: {
		Code:    http.StatusUnauthorized,
		Message: "Unauthorized",
	},
	financialErrors.ErrEmptyToken: {
		Code:    http.StatusUnauthorized,
		Message: "Unauthorized",
	},
	financialErrors.ErrInvalidTokenClaims: {
		Code:    http.StatusUnauthorized,
		Message: "Unauthorized",
	},
	financialErrors.ErrInvalidSigningMethod: {
		Code:    http.StatusUnauthorized,
		Message: "Unauthorized",
	},

	// Domain errors
	financialErrors.ErrBudgetNotFound: {
		Code:    http.StatusNotFound,
		Message: "Budget not found",
	},
	financialErrors.ErrCategoryNotFound: {
		Code:    http.StatusNotFound,
		Message: "Category not found",
	},
	financialErrors.ErrCannotBeEmpty: {
		Code:    http.StatusBadRequest,
		Message: "Name cannot be empty",
	},
	financialErrors.ErrInvalidEmail: {
		Code:    http.StatusBadRequest,
		Message: "Invalid email format",
	},
	financialErrors.ErrEmailAlreadyExists: {
		Code:    http.StatusConflict,
		Message: "Email already exists",
	},
	financialErrors.ErrTooLong: {
		Code:    http.StatusBadRequest,
		Message: "Name cannot be more than 255 characters",
	},
	financialErrors.ErrUserNotFound: {
		Code:    http.StatusNotFound,
		Message: "User not found",
	},
	financialErrors.ErrPasswordIsRequired: {
		Code:    http.StatusBadRequest,
		Message: "Password is required",
	},
	financialErrors.ErrCheckHash: {
		Code:    http.StatusBadRequest,
		Message: "Error checking hash",
	},
	financialErrors.ErrNameIsRequired: {
		Code:    http.StatusBadRequest,
		Message: "Name is required",
	},
	financialErrors.ErrUserIDIsRequired: {
		Code:    http.StatusBadRequest,
		Message: "User ID is required",
	},
	financialErrors.ErrSequenceIsRequired: {
		Code:    http.StatusBadRequest,
		Message: "Sequence is required",
	},
	financialErrors.ErrCategoryCycle: {
		Code:    http.StatusBadRequest,
		Message: "Category cannot be its own parent or create a cycle",
	},
	financialErrors.ErrBudgetInvalidTotal: {
		Code:    http.StatusBadRequest,
		Message: "Sum of budget item percentages must equal 100%",
	},
}

func GetResponseError(err error) *ResponseError {
	// Try direct match first
	if responseError, ok := responseErrors[err]; ok {
		return responseError
	}

	// Try using errors.Is for wrapped errors
	for knownErr, responseError := range responseErrors {
		if errors.Is(err, knownErr) {
			return responseError
		}
	}

	return NewResponseError(http.StatusInternalServerError, "Internal Server Error")
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewResponseError(code int, message string) *ResponseError {
	return &ResponseError{
		Code:    code,
		Message: message,
	}
}
