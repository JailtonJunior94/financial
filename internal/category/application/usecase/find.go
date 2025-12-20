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
		o11y       o11y.Telemetry
		repository interfaces.CategoryRepository
	}
)

func NewFindCategoryUseCase(
	o11y o11y.Telemetry,
	repository interfaces.CategoryRepository,
) FindCategoryUseCase {
	return &findCategoryUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *findCategoryUseCase) Execute(ctx context.Context, userID string) ([]*dtos.CategoryOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "find_category_usecase.execute")
	defer span.End()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		span.AddEvent(
			"error parsing user id",
			o11y.Attribute{Key: "user_id", Value: userID},
			o11y.Attribute{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, err, "error parsing user id", o11y.Field{Key: "user_id", Value: userID})
		return nil, err
	}

	categories, err := u.repository.List(ctx, user)
	if err != nil {
		span.AddEvent(
			"error listing categories from repository",
			o11y.Attribute{Key: "user_id", Value: userID},
			o11y.Attribute{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, err, "error listing categories from repository", o11y.Field{Key: "user_id", Value: userID})
		return nil, err
	}

	categoriesOutput := linq.Map(categories, func(category *entities.Category) *dtos.CategoryOutput {
		return &dtos.CategoryOutput{
			ID:        category.ID.String(),
			Name:      category.Name.String(),
			Sequence:  category.Sequence.Value(),
			CreatedAt: category.CreatedAt.Value(),
		}
	})

	return categoriesOutput, nil
}
