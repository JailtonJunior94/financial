package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/o11y"
)

type (
	CreateCategoryUseCase interface {
		Execute(ctx context.Context, userID string, input *CreateCategoryInput) (*CreateCategoryOutput, error)
	}

	createCategoryUseCase struct {
		o11y       o11y.Observability
		repository interfaces.CategoryRepository
	}
)

func NewCreateCategoryUseCase(
	o11y o11y.Observability,
	repository interfaces.CategoryRepository,
) CreateCategoryUseCase {
	return &createCategoryUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *createCategoryUseCase) Execute(ctx context.Context, userID string, input *CreateCategoryInput) (*CreateCategoryOutput, error) {
	ctx, span := u.o11y.Start(ctx, "create_category_usecase.execute")
	defer span.End()

	newCategory, err := entities.NewCategory(userID, input.Name, input.Sequence)
	if err != nil {
		span.AddStatus(o11y.Error, "error parsing category")
		span.AddAttributes(o11y.Attributes{Key: "user_id", Value: userID})
		return nil, err
	}

	category, err := u.repository.Insert(ctx, newCategory)
	if err != nil {
		span.AddStatus(o11y.Error, "error creating category")
		span.AddAttributes(
			o11y.Attributes{Key: "user_id", Value: userID},
			o11y.Attributes{Key: "error", Value: err},
		)
		return nil, err
	}

	return &CreateCategoryOutput{
		ID:        category.ID,
		Name:      category.Name,
		Sequence:  category.Sequence,
		CreatedAt: category.CreatedAt,
	}, nil
}
