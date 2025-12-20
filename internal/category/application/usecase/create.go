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
		Execute(ctx context.Context, userID string, input *dtos.CategoryInput) (*dtos.CategoryOutput, error)
	}

	createCategoryUseCase struct {
		o11y       o11y.Telemetry
		repository interfaces.CategoryRepository
	}
)

func NewCreateCategoryUseCase(
	o11y o11y.Telemetry,
	repository interfaces.CategoryRepository,
) CreateCategoryUseCase {
	return &createCategoryUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *createCategoryUseCase) Execute(ctx context.Context, userID string, input *dtos.CategoryInput) (*dtos.CategoryOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "create_category_usecase.execute")
	defer span.End()

	category, err := factories.CreateCategory(userID, input.ParentID, input.Name, input.Sequence)
	if err != nil {
		span.AddEvent(
			"error creating category entity",
			o11y.Attribute{Key: "user_id", Value: userID},
			o11y.Attribute{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, err, "error creating category entity", o11y.Field{Key: "user_id", Value: userID})
		return nil, err
	}

	if err := u.repository.Save(ctx, category); err != nil {
		span.AddEvent(
			"error saving category to repository",
			o11y.Attribute{Key: "user_id", Value: userID},
			o11y.Attribute{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, err, "error saving category to repository", o11y.Field{Key: "user_id", Value: userID})
		return nil, err
	}

	return &dtos.CategoryOutput{
		ID:        category.ID.String(),
		Name:      category.Name.String(),
		Sequence:  category.Sequence.Value(),
		CreatedAt: category.CreatedAt.Value(),
	}, nil
}
