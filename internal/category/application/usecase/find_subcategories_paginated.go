package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	categorydomain "github.com/jailtonjunior94/financial/internal/category/domain"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
	"github.com/jailtonjunior94/financial/pkg/pagination"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

const (
	defaultSubcategoryLimit = 20
	maxSubcategoryLimit     = 100
)

type findSubcategoriesPaginatedUseCase struct {
	o11y            observability.Observability
	fm              *metrics.FinancialMetrics
	categoryRepo    interfaces.CategoryRepository
	subcategoryRepo interfaces.SubcategoryRepository
}

func NewFindSubcategoriesPaginatedUseCase(
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
	categoryRepo interfaces.CategoryRepository,
	subcategoryRepo interfaces.SubcategoryRepository,
) FindSubcategoriesPaginatedUseCase {
	return &findSubcategoriesPaginatedUseCase{
		o11y:            o11y,
		fm:              fm,
		categoryRepo:    categoryRepo,
		subcategoryRepo: subcategoryRepo,
	}
}

func (u *findSubcategoriesPaginatedUseCase) Execute(ctx context.Context, userID, categoryID string, limit int, cursor string) (*dtos.SubcategoryPaginatedOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "find_subcategories_paginated_usecase.execute")
	defer span.End()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return nil, err
	}

	catID, err := vos.NewUUIDFromString(categoryID)
	if err != nil {
		return nil, err
	}

	category, err := u.categoryRepo.FindByID(ctx, user, catID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, categorydomain.ErrCategoryNotFound
	}

	if limit <= 0 {
		limit = defaultSubcategoryLimit
	}
	if limit > maxSubcategoryLimit {
		limit = maxSubcategoryLimit
	}

	decodedCursor, err := pagination.DecodeCursor(cursor)
	if err != nil {
		return nil, err
	}

	subcategories, err := u.subcategoryRepo.ListPaginated(ctx, interfaces.ListSubcategoriesParams{
		UserID:     user,
		CategoryID: catID,
		Limit:      limit + 1,
		Cursor:     decodedCursor,
	})
	if err != nil {
		return nil, err
	}

	hasNext := len(subcategories) > limit
	if hasNext {
		subcategories = subcategories[:limit]
	}

	var nextCursor *string
	if hasNext && len(subcategories) > 0 {
		last := subcategories[len(subcategories)-1]
		newCursor := pagination.Cursor{
			Fields: map[string]interface{}{
				"sequence": last.Sequence.Value(),
				"id":       last.ID.String(),
			},
		}
		encoded, err := pagination.EncodeCursor(newCursor)
		if err != nil {
			return nil, err
		}
		nextCursor = &encoded
	}

	data := make([]dtos.SubcategoryOutput, 0, len(subcategories))
	for _, s := range subcategories {
		data = append(data, dtos.SubcategoryOutput{
			ID:         s.ID.String(),
			CategoryID: s.CategoryID.String(),
			Name:       s.Name.String(),
			Sequence:   s.Sequence.Value(),
			CreatedAt:  s.CreatedAt.ValueOr(time.Time{}),
		})
	}

	return &dtos.SubcategoryPaginatedOutput{
		Data: data,
		Pagination: dtos.CategoryPaginationMeta{
			Limit:      limit,
			HasNext:    hasNext,
			NextCursor: nextCursor,
		},
	}, nil
}
