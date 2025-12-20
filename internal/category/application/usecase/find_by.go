package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/linq"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	FindCategoryByUseCase interface {
		Execute(ctx context.Context, userID, id string) (*dtos.CategoryOutput, error)
	}

	findCategoryByUseCase struct {
		o11y       o11y.Telemetry
		repository interfaces.CategoryRepository
	}
)

func NewFindCategoryByUseCase(
	o11y o11y.Telemetry,
	repository interfaces.CategoryRepository,
) FindCategoryByUseCase {
	return &findCategoryByUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *findCategoryByUseCase) Execute(ctx context.Context, userID, id string) (*dtos.CategoryOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "find_category_by_usecase.execute")
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

	categoryID, err := vos.NewUUIDFromString(id)
	if err != nil {
		span.AddEvent(
			"error parsing category id",
			o11y.Attribute{Key: "category_id", Value: id},
			o11y.Attribute{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, err, "error parsing category id", o11y.Field{Key: "category_id", Value: id})
		return nil, err
	}

	category, err := u.repository.FindByID(ctx, user, categoryID)
	if err != nil {
		span.AddEvent(
			"error finding category by id",
			o11y.Attribute{Key: "user_id", Value: userID},
			o11y.Attribute{Key: "category_id", Value: id},
			o11y.Attribute{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, err, "error finding category by id",
			o11y.Field{Key: "user_id", Value: userID},
			o11y.Field{Key: "category_id", Value: id})
		return nil, err
	}

	if category == nil {
		span.AddEvent(
			"category not found",
			o11y.Attribute{Key: "user_id", Value: userID},
			o11y.Attribute{Key: "category_id", Value: id},
		)
		u.o11y.Logger().Error(ctx, customErrors.ErrCategoryNotFound, "category not found",
			o11y.Field{Key: "user_id", Value: userID},
			o11y.Field{Key: "category_id", Value: id})
		return nil, customErrors.ErrCategoryNotFound
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
