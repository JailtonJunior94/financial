package usecase

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type removeSubcategoryUseCase struct {
	o11y            observability.Observability
	fm              *metrics.FinancialMetrics
	categoryRepo    interfaces.CategoryRepository
	subcategoryRepo interfaces.SubcategoryRepository
}

func NewRemoveSubcategoryUseCase(
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
	categoryRepo interfaces.CategoryRepository,
	subcategoryRepo interfaces.SubcategoryRepository,
) RemoveSubcategoryUseCase {
	return &removeSubcategoryUseCase{
		o11y:            o11y,
		fm:              fm,
		categoryRepo:    categoryRepo,
		subcategoryRepo: subcategoryRepo,
	}
}

func (u *removeSubcategoryUseCase) Execute(ctx context.Context, userID, categoryID, id string) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "remove_subcategory_usecase.execute")
	defer span.End()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return err
	}

	catID, err := vos.NewUUIDFromString(categoryID)
	if err != nil {
		return err
	}

	subID, err := vos.NewUUIDFromString(id)
	if err != nil {
		return err
	}

	category, err := u.categoryRepo.FindByID(ctx, user, catID)
	if err != nil {
		return err
	}
	if category == nil {
		return customErrors.ErrCategoryNotFound
	}

	subcategory, err := u.subcategoryRepo.FindByID(ctx, user, catID, subID)
	if err != nil {
		return err
	}
	if subcategory == nil {
		return customErrors.ErrSubcategoryNotFound
	}

	return u.subcategoryRepo.SoftDelete(ctx, subcategory.ID)
}
