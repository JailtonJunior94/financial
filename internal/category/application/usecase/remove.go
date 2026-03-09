package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	"github.com/jailtonjunior94/financial/internal/category/infrastructure/repositories"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	RemoveCategoryUseCase interface {
		Execute(ctx context.Context, userID, id string) error
	}

	removeCategoryUseCase struct {
		o11y         observability.Observability
		fm           *metrics.FinancialMetrics
		uow          uow.UnitOfWork
		categoryRepo interfaces.CategoryRepository
	}
)

func NewRemoveCategoryUseCase(
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
	unitOfWork uow.UnitOfWork,
	categoryRepo interfaces.CategoryRepository,
) RemoveCategoryUseCase {
	return &removeCategoryUseCase{
		o11y:         o11y,
		fm:           fm,
		uow:          unitOfWork,
		categoryRepo: categoryRepo,
	}
}

func (u *removeCategoryUseCase) Execute(ctx context.Context, userID, id string) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "remove_category_usecase.execute")
	defer span.End()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return err
	}

	categoryID, err := vos.NewUUIDFromString(id)
	if err != nil {
		return err
	}

	category, err := u.categoryRepo.FindByID(ctx, user, categoryID)
	if err != nil {
		return err
	}

	if category == nil {
		return customErrors.ErrCategoryNotFound
	}

	return u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		subcatRepo := repositories.NewSubcategoryRepository(tx, u.o11y, u.fm)
		catRepo := repositories.NewCategoryRepository(tx, u.o11y, u.fm)

		if err := subcatRepo.SoftDeleteByCategoryID(ctx, category.ID); err != nil {
			return err
		}

		return catRepo.SoftDelete(ctx, category.ID)
	})
}
