package domain

import (
	"net/http"

	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
)

// ErrorMappings returns the HTTP status mappings for category domain errors.
// Call NewErrorHandler(o11y, domain.ErrorMappings()) in the category module.
func ErrorMappings() map[error]httperrors.ErrorMapping {
	return map[error]httperrors.ErrorMapping{
		ErrCategoryNotFound: {
			Status:  http.StatusNotFound,
			Message: "Category not found",
		},
	}
}
