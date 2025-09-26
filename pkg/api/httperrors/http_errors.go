package httperrors

import (
	"net/http"

	financialErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
)

var responseErrors = map[error]*ResponseError{
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
}

func GetResponseError(err error) *ResponseError {
	if responseError, ok := responseErrors[err]; ok {
		return responseError
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
