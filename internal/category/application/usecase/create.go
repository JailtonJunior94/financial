package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/domain/factories"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	CreateCategoryUseCase interface {
		Execute(ctx context.Context, userID string, input *dtos.CategoryInput) (*dtos.CategoryOutput, error)
	}

	createCategoryUseCase struct {
		o11y       observability.Observability
		repository interfaces.CategoryRepository
	}
)

func NewCreateCategoryUseCase(
	o11y observability.Observability,
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

	// Validate user_id
	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		span.AddEvent(
			"error parsing user id",
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "error", Value: err},
		)
		return nil, err
	}

	// Validate parent_id exists if provided
	if input.ParentID != "" {
		u.o11y.Logger().Info(ctx, "validating parent category",
			observability.String("user_id", userID),
			observability.String("parent_id", input.ParentID),
			observability.String("category_name", input.Name))

		parentID, err := vos.NewUUIDFromString(input.ParentID)
		if err != nil {
			span.AddEvent(
				"error parsing parent id",
				observability.Field{Key: "parent_id", Value: input.ParentID},
				observability.Field{Key: "error", Value: err},
			)
			u.o11y.Logger().Error(ctx, "invalid parent_id format",
				observability.Error(err),
				observability.String("user_id", userID),
				observability.String("parent_id", input.ParentID))
			return nil, err
		}

		// Check if parent category exists
		parentCategory, err := u.repository.FindByID(ctx, user, parentID)
		if err != nil {
			span.AddEvent(
				"error finding parent category",
				observability.Field{Key: "user_id", Value: userID},
				observability.Field{Key: "parent_id", Value: input.ParentID},
				observability.Field{Key: "error", Value: err},
			)
			u.o11y.Logger().Error(ctx, "database error while finding parent category",
				observability.Error(err),
				observability.String("user_id", userID),
				observability.String("parent_id", input.ParentID))
			return nil, err
		}

		if parentCategory == nil {
			span.AddEvent(
				"parent category not found",
				observability.Field{Key: "user_id", Value: userID},
				observability.Field{Key: "parent_id", Value: input.ParentID},
			)
			u.o11y.Logger().Error(ctx, "parent category does not exist or belongs to different user",
				observability.Error(customErrors.ErrCategoryNotFound),
				observability.String("user_id", userID),
				observability.String("parent_id", input.ParentID),
				observability.String("requested_child_name", input.Name))
			return nil, customErrors.ErrCategoryNotFound
		}

		u.o11y.Logger().Info(ctx, "parent category validated successfully",
			observability.String("user_id", userID),
			observability.String("parent_id", input.ParentID),
			observability.String("parent_name", parentCategory.Name.String()))
	}

	category, err := factories.CreateCategory(userID, input.ParentID, input.Name, input.Sequence)
	if err != nil {
		span.AddEvent(
			"error creating category entity",
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "error", Value: err},
		)

		return nil, err
	}

	if err := u.repository.Save(ctx, category); err != nil {
		span.AddEvent(
			"error saving category to repository",
			observability.Field{Key: "user_id", Value: userID},
			observability.Field{Key: "error", Value: err},
		)

		return nil, err
	}

	return &dtos.CategoryOutput{
		ID:        category.ID.String(),
		Name:      category.Name.String(),
		Sequence:  category.Sequence.Value(),
		CreatedAt: category.CreatedAt.ValueOr(time.Time{}),
	}, nil
}
