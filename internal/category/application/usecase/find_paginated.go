package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
	"github.com/jailtonjunior94/financial/pkg/pagination"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	FindCategoryPaginatedUseCase interface {
		Execute(ctx context.Context, input FindCategoryPaginatedInput) (*FindCategoryPaginatedOutput, error)
	}

	FindCategoryPaginatedInput struct {
		UserID string
		Limit  int
		Cursor string
	}

	FindCategoryPaginatedOutput struct {
		Categories []*dtos.CategoryOutput
		NextCursor *string
	}

	findCategoryPaginatedUseCase struct {
		o11y       observability.Observability
		fm         *metrics.FinancialMetrics
		repository interfaces.CategoryRepository
	}
)

func NewFindCategoryPaginatedUseCase(
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
	repository interfaces.CategoryRepository,
) FindCategoryPaginatedUseCase {
	return &findCategoryPaginatedUseCase{
		o11y:       o11y,
		fm:         fm,
		repository: repository,
	}
}

func (u *findCategoryPaginatedUseCase) Execute(ctx context.Context, input FindCategoryPaginatedInput) (*FindCategoryPaginatedOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "find_category_paginated_usecase.execute")
	defer span.End()

	userID, err := vos.NewUUIDFromString(input.UserID)
	if err != nil {
		return nil, err
	}

	cursor, err := pagination.DecodeCursor(input.Cursor)
	if err != nil {
		return nil, err
	}

	categories, err := u.repository.ListPaginated(ctx, interfaces.ListCategoriesParams{
		UserID: userID,
		Limit:  input.Limit + 1,
		Cursor: cursor,
	})
	if err != nil {
		return nil, err
	}

	hasNext := len(categories) > input.Limit
	if hasNext {
		categories = categories[:input.Limit]
	}

	var nextCursor *string
	if hasNext && len(categories) > 0 {
		lastCategory := categories[len(categories)-1]
		newCursor := pagination.Cursor{
			Fields: map[string]interface{}{
				"sequence": lastCategory.Sequence.Value(),
				"id":       lastCategory.ID.String(),
			},
		}
		encoded, err := pagination.EncodeCursor(newCursor)
		if err != nil {
			return nil, err
		}
		nextCursor = &encoded
	}

	output := make([]*dtos.CategoryOutput, len(categories))
	for i, category := range categories {
		output[i] = &dtos.CategoryOutput{
			ID:        category.ID.String(),
			Name:      category.Name.String(),
			Sequence:  category.Sequence.Value(),
			CreatedAt: category.CreatedAt.ValueOr(time.Time{}),
		}
	}

	return &FindCategoryPaginatedOutput{
		Categories: output,
		NextCursor: nextCursor,
	}, nil
}
