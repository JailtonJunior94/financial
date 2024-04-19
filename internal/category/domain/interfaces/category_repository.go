package interfaces

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
)

type CategoryRepository interface {
	Find(ctx context.Context, userID string) ([]*entities.Category, error)
	FindByID(ctx context.Context, userID, id string) (*entities.Category, error)
	Insert(ctx context.Context, category *entities.Category) (*entities.Category, error)
	Update(ctx context.Context, category *entities.Category) (*entities.Category, error)
}
