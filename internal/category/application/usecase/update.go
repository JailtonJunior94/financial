package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	UpdateCategoryUseCase interface {
		Execute(ctx context.Context, userID, id string, input *dtos.CategoryInput) (*dtos.CategoryOutput, error)
	}

	updateCategoryUseCase struct {
		o11y       o11y.Telemetry
		repository interfaces.CategoryRepository
	}
)

func NewUpdateCategoryUseCase(
	o11y o11y.Telemetry,
	repository interfaces.CategoryRepository,
) UpdateCategoryUseCase {
	return &updateCategoryUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *updateCategoryUseCase) Execute(ctx context.Context, userID, id string, input *dtos.CategoryInput) (*dtos.CategoryOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "update_category_usecase.execute")
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

	// Parse and validate parent_id if provided
	var parentID *vos.UUID
	if input.ParentID != "" {
		parsedParentID, err := vos.NewUUIDFromString(input.ParentID)
		if err != nil {
			span.AddEvent(
				"error parsing parent id",
				o11y.Attribute{Key: "parent_id", Value: input.ParentID},
				o11y.Attribute{Key: "error", Value: err},
			)
			u.o11y.Logger().Error(ctx, err, "error parsing parent id", o11y.Field{Key: "parent_id", Value: input.ParentID})
			return nil, err
		}

		// Validate: category cannot be its own parent
		if parsedParentID.String() == categoryID.String() {
			span.AddEvent(
				"category cannot be its own parent",
				o11y.Attribute{Key: "category_id", Value: id},
				o11y.Attribute{Key: "parent_id", Value: input.ParentID},
			)
			u.o11y.Logger().Error(ctx, customErrors.ErrCategoryCycle, "category cannot be its own parent",
				o11y.Field{Key: "category_id", Value: id},
				o11y.Field{Key: "parent_id", Value: input.ParentID})
			return nil, customErrors.ErrCategoryCycle
		}

		// Validate: check for cycles by traversing parent chain
		if err := u.validateNoCycle(ctx, user, parsedParentID, categoryID); err != nil {
			span.AddEvent(
				"cycle detected in category hierarchy",
				o11y.Attribute{Key: "category_id", Value: id},
				o11y.Attribute{Key: "parent_id", Value: input.ParentID},
				o11y.Attribute{Key: "error", Value: err},
			)
			u.o11y.Logger().Error(ctx, err, "cycle detected in category hierarchy",
				o11y.Field{Key: "category_id", Value: id},
				o11y.Field{Key: "parent_id", Value: input.ParentID})
			return nil, err
		}

		parentID = &parsedParentID
	}

	if err := category.Update(input.Name, input.Sequence, parentID); err != nil {
		span.AddEvent(
			"error validating category update",
			o11y.Attribute{Key: "user_id", Value: userID},
			o11y.Attribute{Key: "category_id", Value: id},
			o11y.Attribute{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, err, "error validating category update",
			o11y.Field{Key: "user_id", Value: userID},
			o11y.Field{Key: "category_id", Value: id})
		return nil, err
	}

	if err := u.repository.Update(ctx, category); err != nil {
		span.AddEvent(
			"error updating category in repository",
			o11y.Attribute{Key: "user_id", Value: userID},
			o11y.Attribute{Key: "category_id", Value: id},
			o11y.Attribute{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, err, "error updating category in repository",
			o11y.Field{Key: "user_id", Value: userID},
			o11y.Field{Key: "category_id", Value: id})
		return nil, err
	}

	return &dtos.CategoryOutput{
		ID:       category.ID.String(),
		Name:     category.Name.String(),
		Sequence: category.Sequence.Value(),
	}, nil
}

// validateNoCycle checks if setting parentID as parent would create a cycle
// Uses a recursive CTE query for efficient cycle detection in a single database round-trip
func (u *updateCategoryUseCase) validateNoCycle(ctx context.Context, userID, parentID, categoryID vos.UUID) error {
	cycleExists, err := u.repository.CheckCycleExists(ctx, userID, categoryID, parentID)
	if err != nil {
		return fmt.Errorf("error checking for cycle: %w", err)
	}
	if cycleExists {
		return customErrors.ErrCategoryCycle
	}
	return nil
}
