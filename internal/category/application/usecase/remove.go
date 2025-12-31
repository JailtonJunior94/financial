package usecase

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
)

type (
	RemoveCategoryUseCase interface {
		Execute(ctx context.Context, userID, id string) error
	}

	removeCategoryUseCase struct {
		o11y       observability.Observability
		repository interfaces.CategoryRepository
	}
)

func NewRemoveCategoryUseCase(
	o11y observability.Observability,
	repository interfaces.CategoryRepository,
) RemoveCategoryUseCase {
	return &removeCategoryUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *removeCategoryUseCase) Execute(ctx context.Context, userID, id string) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "remove_category_usecase.execute")
	defer span.End()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		span.AddEvent(
			"error parsing user id",
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "error", Value: err},
		)

		return err
	}

	categoryID, err := vos.NewUUIDFromString(id)
	if err != nil {
		span.AddEvent(
			"error parsing category id",
			observability.Field{Key: "category_id", Value: id},
			observability.Field{Key: "error", Value: err},
		)

		return err
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
		return err
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
		return customErrors.ErrCategoryNotFound
	}

	if err := u.repository.Update(ctx, category.Delete()); err != nil {
		span.AddEvent(
			"error deleting category in repository",
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "category_id", Value: id},
			observability.Field{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, "error deleting category in repository",
			observability.Error(err),
			observability.String("user_id", userID),
			observability.String("category_id", id))
		return err
	}

	return nil
}
