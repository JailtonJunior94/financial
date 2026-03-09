package domain

import (
	"net/http"

	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
)

// ErrorMappings returns the HTTP status mappings for user domain errors.
// Call NewErrorHandler(o11y, domain.ErrorMappings()) in the user module.
func ErrorMappings() map[error]httperrors.ErrorMapping {
	return map[error]httperrors.ErrorMapping{
		ErrUserNotFound: {
			Status:  http.StatusNotFound,
			Message: "User not found",
		},
		ErrEmailAlreadyExists: {
			Status:  http.StatusConflict,
			Message: "Email already exists",
		},
	}
}
