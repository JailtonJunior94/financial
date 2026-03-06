package interfaces

import (
"context"

"github.com/jailtonjunior94/financial/internal/category/domain/entities"
"github.com/jailtonjunior94/financial/pkg/pagination"

"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type ListSubcategoriesParams struct {
UserID     vos.UUID
CategoryID vos.UUID
Limit      int
Cursor     pagination.Cursor
}

type SubcategoryRepository interface {
FindByID(ctx context.Context, userID, categoryID, id vos.UUID) (*entities.Subcategory, error)
FindByCategoryID(ctx context.Context, categoryID vos.UUID) ([]*entities.Subcategory, error)
ListPaginated(ctx context.Context, params ListSubcategoriesParams) ([]*entities.Subcategory, error)
Save(ctx context.Context, subcategory *entities.Subcategory) error
Update(ctx context.Context, subcategory *entities.Subcategory) error
SoftDelete(ctx context.Context, id vos.UUID) error
SoftDeleteByCategoryID(ctx context.Context, categoryID vos.UUID) error
}
