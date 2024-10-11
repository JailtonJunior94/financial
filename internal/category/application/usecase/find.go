package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/linq"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	FindCategoryUseCase interface {
		Execute(ctx context.Context, userID string) ([]*dtos.CategoryOutput, error)
	}

	findCategoryUseCase struct {
		o11y       o11y.Observability
		repository interfaces.CategoryRepository
	}
)

func NewFindCategoryUseCase(
	o11y o11y.Observability,
	repository interfaces.CategoryRepository,
) FindCategoryUseCase {
	return &findCategoryUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *findCategoryUseCase) Execute(ctx context.Context, userID string) ([]*dtos.CategoryOutput, error) {
	ctx, span := u.o11y.Start(ctx, "find_category_usecase.execute")
	defer span.End()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, err.Error(), o11y.Attributes{Key: "error", Value: err})
		return nil, err
	}

	categories, err := u.repository.Find(ctx, user)
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, err.Error(), o11y.Attributes{Key: "error", Value: err})
		return nil, err
	}

	categoriesOutput := linq.Map(categories, func(category *entities.Category) *dtos.CategoryOutput {
		return &dtos.CategoryOutput{
			ID:        category.ID.String(),
			Name:      category.Name,
			Sequence:  category.Sequence,
			CreatedAt: category.CreatedAt,
		}
	})

	return categoriesOutput, nil
}
