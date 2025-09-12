package interfaces

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/category/domain/entities"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type CategoryRepository interface {
	List(ctx context.Context, userID vos.UUID) ([]*entities.Category, error)
	FindByID(ctx context.Context, userID, id vos.UUID) (*entities.Category, error)
	Save(ctx context.Context, category *entities.Category) error
	Update(ctx context.Context, category *entities.Category) error
}
