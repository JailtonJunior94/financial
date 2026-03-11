package user

import (
	"net/http"

	"github.com/jailtonjunior94/financial/internal/user/domain"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
)

// ErrorMappings returns the HTTP status mappings for user domain errors.
func ErrorMappings() map[error]httperrors.ErrorMapping {
	return map[error]httperrors.ErrorMapping{
		domain.ErrUserNotFound: {
			Status:  http.StatusNotFound,
			Message: "User not found",
		},
		domain.ErrEmailAlreadyExists: {
			Status:  http.StatusConflict,
			Message: "Email already exists",
		},
	}
}
