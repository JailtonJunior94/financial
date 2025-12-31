package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/linq"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	FindCategoryByUseCase interface {
		Execute(ctx context.Context, userID, id string) (*dtos.CategoryOutput, error)
	}

	findCategoryByUseCase struct {
		o11y       observability.Observability
		repository interfaces.CategoryRepository
	}
)

func NewFindCategoryByUseCase(
	o11y observability.Observability,
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
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "error", Value: err},
		)

		return nil, err
	}

	categoryID, err := vos.NewUUIDFromString(id)
	if err != nil {
		span.AddEvent(
			"error parsing category id",
			observability.Field{Key: "category_id", Value: id},
			observability.Field{Key: "error", Value: err},
		)

		return nil, err
	}

	category, err := u.repository.FindByID(ctx, user, categoryID)
	if err != nil {
		span.AddEvent(
			"error finding category by id",
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "category_id", Value: id},
			observability.Field{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, "error finding category by id",
			observability.Error(err),
			observability.String("user_id", userID),
			observability.String("category_id", id))
		return nil, err
	}

	if category == nil {
		span.AddEvent(
			"category not found",
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "category_id", Value: id},
		)
		u.o11y.Logger().Error(ctx, "category not found",
			observability.Error(customErrors.ErrCategoryNotFound),
			observability.String("user_id", userID),
			observability.String("category_id", id))
		return nil, customErrors.ErrCategoryNotFound
	}

	return &dtos.CategoryOutput{
		ID:        category.ID.String(),
		Name:      category.Name.String(),
		Sequence:  category.Sequence.Value(),
		CreatedAt: category.CreatedAt.ValueOr(time.Time{}),
		Children: linq.Map(category.Children, func(child entities.Category) dtos.CategoryOutput {
			return dtos.CategoryOutput{
				ID:        child.ID.String(),
				Name:      child.Name.String(),
				Sequence:  child.Sequence.Value(),
				CreatedAt: child.CreatedAt.ValueOr(time.Time{}),
			}
		}),
	}, nil

}
