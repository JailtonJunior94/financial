package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	UpdateCategoryUseCase interface {
		Execute(ctx context.Context, userID, id string, input *dtos.CategoryInput) (*dtos.CategoryOutput, error)
	}

	updateCategoryUseCase struct {
		o11y       o11y.Observability
		repository interfaces.CategoryRepository
	}
)

func NewUpdateCategoryUseCase(
	o11y o11y.Observability,
	repository interfaces.CategoryRepository,
) UpdateCategoryUseCase {
	return &updateCategoryUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *updateCategoryUseCase) Execute(ctx context.Context, userID, id string, input *dtos.CategoryInput) (*dtos.CategoryOutput, error) {
	ctx, span := u.o11y.Start(ctx, "Update_category_usecase.execute")
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

	if err := u.repository.Update(ctx, category.Update(input.Name, input.Sequence)); err != nil {
		span.AddAttributes(
			ctx, o11y.Error, "error updating category",
			o11y.Attributes{Key: "user_id", Value: userID},
			o11y.Attributes{Key: "error", Value: err},
		)
		return nil, err
	}

	return &dtos.CategoryOutput{
		ID:       category.ID.String(),
		Name:     category.Name.String(),
		Sequence: category.Sequence.Value(),
	}, nil
}
