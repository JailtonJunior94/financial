package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
)

type CreateSubcategoryUseCase interface {
	Execute(ctx context.Context, userID, categoryID string, input *dtos.SubcategoryInput) (*dtos.SubcategoryOutput, error)
}

type FindSubcategoryByUseCase interface {
	Execute(ctx context.Context, userID, categoryID, id string) (*dtos.SubcategoryOutput, error)
}

type FindSubcategoriesPaginatedUseCase interface {
	Execute(ctx context.Context, userID, categoryID string, limit int, cursor string) (*dtos.SubcategoryPaginatedOutput, error)
}

type UpdateSubcategoryUseCase interface {
	Execute(ctx context.Context, userID, categoryID, id string, input *dtos.SubcategoryInput) (*dtos.SubcategoryOutput, error)
}

type RemoveSubcategoryUseCase interface {
	Execute(ctx context.Context, userID, categoryID, id string) error
}
