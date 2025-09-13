package usecase

import (
	"context"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
)

type (
	RemoveCategoryUseCase interface {
		Execute(ctx context.Context, userID, id string) error
	}

	removeCategoryUseCase struct {
		o11y       o11y.Observability
		repository interfaces.CategoryRepository
	}
)

func NewRemoveCategoryUseCase(
	o11y o11y.Observability,
	repository interfaces.CategoryRepository,
) *removeCategoryUseCase {
	return &removeCategoryUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

func (u *removeCategoryUseCase) Execute(ctx context.Context, userID, id string) error {
	ctx, span := u.o11y.Start(ctx, "remove_category_usecase.execute")
	defer span.End()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, err.Error(), o11y.Attributes{Key: "error", Value: err})
		return err
	}

	categoryID, err := vos.NewUUIDFromString(id)
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, err.Error(), o11y.Attributes{Key: "error", Value: err})
		return err
	}

	category, err := u.repository.FindByID(ctx, user, categoryID)
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, err.Error(), o11y.Attributes{Key: "error", Value: err})
		return err
	}

	if err := u.repository.Update(ctx, category.Delete()); err != nil {
		span.AddAttributes(
			ctx, o11y.Error, "error updating category",
			o11y.Attributes{Key: "user_id", Value: userID},
			o11y.Attributes{Key: "error", Value: err},
		)
		return err
	}

	return nil
}
