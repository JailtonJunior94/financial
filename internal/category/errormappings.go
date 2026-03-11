package category

import (
	"net/http"

	"github.com/jailtonjunior94/financial/internal/category/domain"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
)

// ErrorMappings returns the HTTP status mappings for category domain errors.
func ErrorMappings() map[error]httperrors.ErrorMapping {
	return map[error]httperrors.ErrorMapping{
		domain.ErrCategoryNotFound: {
			Status:  http.StatusNotFound,
			Message: "Category not found",
		},
	}
}
