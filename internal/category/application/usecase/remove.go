package usecase

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
)

type (
	RemoveCategoryUseCase interface {
		Execute(ctx context.Context, userID, id string) error
	}

	removeCategoryUseCase struct {
		o11y       o11y.Telemetry
		repository interfaces.CategoryRepository
	}
)

func NewRemoveCategoryUseCase(
	o11y o11y.Telemetry,
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
			o11y.Attribute{Key: "user_id", Value: userID},
			o11y.Attribute{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, err, "error parsing user id", o11y.Field{Key: "user_id", Value: userID})
		return err
	}

	categoryID, err := vos.NewUUIDFromString(id)
	if err != nil {
		span.AddEvent(
			"error parsing category id",
			o11y.Attribute{Key: "category_id", Value: id},
			o11y.Attribute{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, err, "error parsing category id", o11y.Field{Key: "category_id", Value: id})
		return err
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
		return err
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
		return customErrors.ErrCategoryNotFound
	}

	if err := u.repository.Update(ctx, category.Delete()); err != nil {
		span.AddEvent(
			"error deleting category in repository",
			o11y.Attribute{Key: "user_id", Value: userID},
			o11y.Attribute{Key: "category_id", Value: id},
			o11y.Attribute{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, err, "error deleting category in repository",
			o11y.Field{Key: "user_id", Value: userID},
			o11y.Field{Key: "category_id", Value: id})
		return err
	}

	return nil
}
