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
	FindCategoryByUseCase interface {
		Execute(ctx context.Context, userID, id string) (*dtos.CategoryOutput, error)
	}

	findCategoryByUseCase struct {
		o11y       o11y.Observability
		repository interfaces.CategoryRepository
	}
)

func NewFindCategoryByUseCase(
	o11y o11y.Observability,
	repository interfaces.CategoryRepository,
) FindCategoryByUseCase {
	return &findCategoryByUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *findCategoryByUseCase) Execute(ctx context.Context, userID, id string) (*dtos.CategoryOutput, error) {
	ctx, span := u.o11y.Start(ctx, "find_category_by_usecase.execute")
	defer span.End()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, err.Error(), o11y.Attributes{Key: "error", Value: err})
		return nil, err
	}

	categoryID, err := vos.NewUUIDFromString(id)
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, err.Error(), o11y.Attributes{Key: "error", Value: err})
		return nil, err
	}

	category, err := u.repository.FindByID(ctx, user, categoryID)
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, err.Error(), o11y.Attributes{Key: "error", Value: err})
		return nil, err
	}

	return &dtos.CategoryOutput{
		ID:        category.ID.String(),
		Name:      category.Name.String(),
		Sequence:  category.Sequence.Value(),
		CreatedAt: category.CreatedAt.Value(),
		Children: linq.Map(category.Children, func(child entities.Category) dtos.CategoryOutput {
			return dtos.CategoryOutput{
				ID:        child.ID.String(),
				Name:      child.Name.String(),
				Sequence:  child.Sequence.Value(),
				CreatedAt: child.CreatedAt.Value(),
			}
		}),
	}, nil

}
