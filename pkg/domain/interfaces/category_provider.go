package interfaces

import (
	"context"
	"errors"
)

var (
	ErrCategoryNotFound       = errors.New("category not found")
	ErrCategoryNotOwnedByUser = errors.New("category does not belong to user")
)

// CategoryProvider validates that category IDs exist and belong to the user.
// Shared interface between the budget and category modules (Port & Adapter).
type CategoryProvider interface {
	ValidateCategories(ctx context.Context, userID string, categoryIDs []string) error
}
