package usecase

import (
	"context"

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

	if err := u.repository.Update(ctx, category.Update(input.Name, input.Sequence, parentID)); err != nil {
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
// by traversing the parent chain and ensuring we don't encounter categoryID
func (u *updateCategoryUseCase) validateNoCycle(ctx context.Context, userID, parentID, categoryID vos.UUID) error {
	currentID := parentID
	visited := make(map[string]bool)

	// Traverse the parent chain upwards
	for {
		// Check if we've seen this ID before (cycle detection)
		if visited[currentID.String()] {
			return customErrors.ErrCategoryCycle
		}
		visited[currentID.String()] = true

		// Check if we've reached the category being updated (cycle!)
		if currentID.String() == categoryID.String() {
			return customErrors.ErrCategoryCycle
		}

		// Fetch the current category to get its parent
		parent, err := u.repository.FindByID(ctx, userID, currentID)
		if err != nil {
			return err
		}

		// If no parent found or parent has no parent_id, we've reached the top
		if parent == nil || parent.ParentID == nil {
			break
		}

		// Move to the next parent in the chain
		currentID = *parent.ParentID
	}

	return nil
}
