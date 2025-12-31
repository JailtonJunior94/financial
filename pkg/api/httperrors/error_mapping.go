package httperrors

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	customerrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
)

// ErrorMapping represents the HTTP status code and message for an error
type ErrorMapping struct {
	Status  int
	Message string
}

// ErrorMapper maps domain errors to HTTP status codes and messages
type ErrorMapper interface {
	MapError(err error) ErrorMapping
}

type errorMapper struct {
	domainMappings map[error]ErrorMapping
}

// NewErrorMapper creates a new error mapper with domain error mappings
func NewErrorMapper() ErrorMapper {
	return &errorMapper{
		domainMappings: buildDomainErrorMappings(),
	}
}

// MapError maps a domain error to an HTTP status code and message
func (m *errorMapper) MapError(err error) ErrorMapping {
	if err == nil {
		return ErrorMapping{
			Status:  http.StatusInternalServerError,
			Message: "Unknown error",
		}
	}

	// 1. Try direct match in the map
	if mapping, ok := m.domainMappings[err]; ok {
		return mapping
	}

	// 2. Try match with errors.Is (for wrapped errors)
	for knownErr, mapping := range m.domainMappings {
		if errors.Is(err, knownErr) {
			return mapping
		}
	}

	// 3. Detect common errors by type
	if isJSONError(err) {
		return ErrorMapping{
			Status:  http.StatusBadRequest,
			Message: "Invalid JSON in request body",
		}
	}

	if isValidationError(err) {
		return ErrorMapping{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		}
	}

	// 4. Default: 500 Internal Server Error
	return ErrorMapping{
		Status:  http.StatusInternalServerError,
		Message: "Internal server error",
	}
}

// buildDomainErrorMappings creates the mapping of domain errors to HTTP status codes
func buildDomainErrorMappings() map[error]ErrorMapping {
	return map[error]ErrorMapping{
		// Validation errors → 400 Bad Request
		customerrors.ErrCannotBeEmpty: {
			Status:  http.StatusBadRequest,
			Message: "Name cannot be empty",
		},
		customerrors.ErrInvalidEmail: {
			Status:  http.StatusBadRequest,
			Message: "Invalid email format",
		},
		customerrors.ErrNameIsRequired: {
			Status:  http.StatusBadRequest,
			Message: "Name is required",
		},
		customerrors.ErrNameCannotBeEmpty: {
			Status:  http.StatusBadRequest,
			Message: "Name cannot be empty",
		},
		customerrors.ErrUserIDIsRequired: {
			Status:  http.StatusBadRequest,
			Message: "User ID is required",
		},
		customerrors.ErrSequenceIsRequired: {
			Status:  http.StatusBadRequest,
			Message: "Sequence is required",
		},
		customerrors.ErrSequenceTooLarge: {
			Status:  http.StatusBadRequest,
			Message: "Sequence cannot be greater than 1000",
		},
		customerrors.ErrPasswordIsRequired: {
			Status:  http.StatusBadRequest,
			Message: "Password is required",
		},
		customerrors.ErrCategoryCycle: {
			Status:  http.StatusBadRequest,
			Message: "Category cannot be its own parent or create a cycle",
		},
		customerrors.ErrBudgetInvalidTotal: {
			Status:  http.StatusBadRequest,
			Message: "Sum of budget item percentages must equal 100%",
		},
		customerrors.ErrNameTooLong: {
			Status:  http.StatusBadRequest,
			Message: "Name cannot be more than 100 characters",
		},
		customerrors.ErrTooLong: {
			Status:  http.StatusBadRequest,
			Message: "Value cannot be more than 255 characters",
		},

		// Not found errors → 404 Not Found
		customerrors.ErrBudgetNotFound: {
			Status:  http.StatusNotFound,
			Message: "Budget not found",
		},
		customerrors.ErrCategoryNotFound: {
			Status:  http.StatusNotFound,
			Message: "Category not found",
		},
		customerrors.ErrUserNotFound: {
			Status:  http.StatusNotFound,
			Message: "User not found",
		},

		// Conflict errors → 409 Conflict
		customerrors.ErrEmailAlreadyExists: {
			Status:  http.StatusConflict,
			Message: "Email already exists",
		},
		customerrors.ErrInvalidParentCategory: {
			Status:  http.StatusConflict,
			Message: "Invalid parent category",
		},

		// Authentication errors → 401 Unauthorized
		customerrors.ErrUnauthorized: {
			Status:  http.StatusUnauthorized,
			Message: "Unauthorized",
		},
		customerrors.ErrMissingAuthHeader: {
			Status:  http.StatusUnauthorized,
			Message: "Missing authorization header",
		},
		customerrors.ErrInvalidAuthFormat: {
			Status:  http.StatusUnauthorized,
			Message: "Invalid authorization format",
		},
		customerrors.ErrInvalidToken: {
			Status:  http.StatusUnauthorized,
			Message: "Invalid token",
		},
		customerrors.ErrTokenExpired: {
			Status:  http.StatusUnauthorized,
			Message: "Token has expired",
		},
		customerrors.ErrEmptyToken: {
			Status:  http.StatusUnauthorized,
			Message: "Empty token provided",
		},
		customerrors.ErrInvalidTokenClaims: {
			Status:  http.StatusUnauthorized,
			Message: "Invalid token claims",
		},
		customerrors.ErrInvalidSigningMethod: {
			Status:  http.StatusUnauthorized,
			Message: "Invalid token signing method",
		},
		customerrors.ErrCheckHash: {
			Status:  http.StatusUnauthorized,
			Message: "Invalid credentials",
		},
	}
}

// isJSONError checks if the error is a JSON parsing error
func isJSONError(err error) bool {
	var syntaxErr *json.SyntaxError
	var unmarshalErr *json.UnmarshalTypeError
	return errors.As(err, &syntaxErr) || errors.As(err, &unmarshalErr)
}

// isValidationError checks if the error is a validation error based on error message
func isValidationError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "invalid") ||
		strings.Contains(errMsg, "required") ||
		strings.Contains(errMsg, "cannot be")
}
