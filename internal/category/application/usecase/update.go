package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	UpdateCategoryUseCase interface {
		Execute(ctx context.Context, userID, id string, input *dtos.CategoryInput) (*dtos.CategoryOutput, error)
	}

	updateCategoryUseCase struct {
		o11y       observability.Observability
		repository interfaces.CategoryRepository
	}
)

func NewUpdateCategoryUseCase(
	o11y observability.Observability,
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
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, "error parsing user id", observability.Error(err), observability.String("user_id", userID))
		return nil, err
	}

	categoryID, err := vos.NewUUIDFromString(id)
	if err != nil {
		span.AddEvent(
			"error parsing category id",
			observability.Field{Key: "category_id", Value: id},
			observability.Field{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, "error parsing category id", observability.Error(err), observability.String("category_id", id))
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

	// Parse and validate parent_id if provided
	var parentID *vos.UUID
	if input.ParentID != "" {
		parsedParentID, err := vos.NewUUIDFromString(input.ParentID)
		if err != nil {
			span.AddEvent(
				"error parsing parent id",
				observability.Field{Key: "parent_id", Value: input.ParentID},
				observability.Field{Key: "error", Value: err},
			)
			u.o11y.Logger().Error(ctx, "error parsing parent id", observability.Error(err), observability.String("parent_id", input.ParentID))
			return nil, err
		}

		// Validate: category cannot be its own parent
		if parsedParentID.String() == categoryID.String() {
			span.AddEvent(
				"category cannot be its own parent",
				observability.Field{Key: "category_id", Value: id},
				observability.Field{Key: "parent_id", Value: input.ParentID},
			)
			u.o11y.Logger().Error(ctx, "category cannot be its own parent",
				observability.Error(customErrors.ErrCategoryCycle),
				observability.String("category_id", id),
				observability.String("parent_id", input.ParentID))
			return nil, customErrors.ErrCategoryCycle
		}

		// Validate: check for cycles by traversing parent chain
		if err := u.validateNoCycle(ctx, user, parsedParentID, categoryID); err != nil {
			span.AddEvent(
				"cycle detected in category hierarchy",
				observability.Field{Key: "category_id", Value: id},
				observability.Field{Key: "parent_id", Value: input.ParentID},
				observability.Field{Key: "error", Value: err},
			)
			u.o11y.Logger().Error(ctx, "cycle detected in category hierarchy",
				observability.Error(err),
				observability.String("category_id", id),
				observability.String("parent_id", input.ParentID))
			return nil, err
		}

		parentID = &parsedParentID
	}

	if err := category.Update(input.Name, input.Sequence, parentID); err != nil {
		span.AddEvent(
			"error validating category update",
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "category_id", Value: id},
			observability.Field{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, "error validating category update",
			observability.Error(err),
			observability.String("user_id", userID),
			observability.String("category_id", id))
		return nil, err
	}

	if err := u.repository.Update(ctx, category); err != nil {
		span.AddEvent(
			"error updating category in repository",
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "category_id", Value: id},
			observability.Field{Key: "error", Value: err},
		)
		u.o11y.Logger().Error(ctx, "error updating category in repository",
			observability.Error(err),
			observability.String("user_id", userID),
			observability.String("category_id", id))
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
