package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type findSubcategoryByUseCase struct {
	o11y            observability.Observability
	fm              *metrics.FinancialMetrics
	categoryRepo    interfaces.CategoryRepository
	subcategoryRepo interfaces.SubcategoryRepository
}

func NewFindSubcategoryByUseCase(
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
	categoryRepo interfaces.CategoryRepository,
	subcategoryRepo interfaces.SubcategoryRepository,
) FindSubcategoryByUseCase {
	return &findSubcategoryByUseCase{
		o11y:            o11y,
		fm:              fm,
		categoryRepo:    categoryRepo,
		subcategoryRepo: subcategoryRepo,
	}
}

func (u *findSubcategoryByUseCase) Execute(ctx context.Context, userID, categoryID, id string) (*dtos.SubcategoryOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "find_subcategory_by_usecase.execute")
	defer span.End()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return nil, err
	}

	catID, err := vos.NewUUIDFromString(categoryID)
	if err != nil {
		return nil, err
	}

	subID, err := vos.NewUUIDFromString(id)
	if err != nil {
		return nil, err
	}

	category, err := u.categoryRepo.FindByID(ctx, user, catID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, customErrors.ErrCategoryNotFound
	}

	subcategory, err := u.subcategoryRepo.FindByID(ctx, user, catID, subID)
	if err != nil {
		return nil, err
	}
	if subcategory == nil {
		return nil, customErrors.ErrSubcategoryNotFound
	}

	return &dtos.SubcategoryOutput{
		ID:         subcategory.ID.String(),
		CategoryID: subcategory.CategoryID.String(),
		Name:       subcategory.Name.String(),
		Sequence:   subcategory.Sequence.Value(),
		CreatedAt:  subcategory.CreatedAt.ValueOr(time.Time{}),
	}, nil
}
