package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/domain/factories"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
)

type (
	CreateCategoryUseCase interface {
		Execute(ctx context.Context, userID string, input *dtos.CreateCategoryInput) (*dtos.CategoryOutput, error)
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

func (u *createCategoryUseCase) Execute(ctx context.Context, userID string, input *dtos.CreateCategoryInput) (*dtos.CategoryOutput, error) {
	ctx, span := u.o11y.Start(ctx, "create_category_usecase.execute")
	defer span.End()

	newCategory, err := factories.CreateCategory(userID, input.ParentID, input.Name, input.Sequence)
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, err.Error(), o11y.Attributes{Key: "error", Value: err})
		return nil, err
	}

	category, err := u.repository.Insert(ctx, newCategory)
	if err != nil {
		span.AddAttributes(
			ctx, o11y.Error, "error creating category",
			o11y.Attributes{Key: "user_id", Value: userID},
			o11y.Attributes{Key: "error", Value: err},
		)
		return nil, err
	}

	return &dtos.CategoryOutput{
		ID:        category.ID.String(),
		Name:      category.Name.String(),
		Sequence:  category.Sequence.Value(),
		CreatedAt: category.CreatedAt.Value(),
	}, nil
}
