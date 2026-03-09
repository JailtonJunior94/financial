package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	categorydomain "github.com/jailtonjunior94/financial/internal/category/domain"
	"github.com/jailtonjunior94/financial/internal/category/domain/factories"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type createSubcategoryUseCase struct {
	o11y            observability.Observability
	fm              *metrics.FinancialMetrics
	categoryRepo    interfaces.CategoryRepository
	subcategoryRepo interfaces.SubcategoryRepository
}

func NewCreateSubcategoryUseCase(
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
	categoryRepo interfaces.CategoryRepository,
	subcategoryRepo interfaces.SubcategoryRepository,
) CreateSubcategoryUseCase {
	return &createSubcategoryUseCase{
		o11y:            o11y,
		fm:              fm,
		categoryRepo:    categoryRepo,
		subcategoryRepo: subcategoryRepo,
	}
}

func (u *createSubcategoryUseCase) Execute(ctx context.Context, userID, categoryID string, input *dtos.SubcategoryInput) (*dtos.SubcategoryOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "create_subcategory_usecase.execute")
	defer span.End()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return nil, err
	}

	catID, err := vos.NewUUIDFromString(categoryID)
	if err != nil {
		return nil, err
	}

	category, err := u.categoryRepo.FindByID(ctx, user, catID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, categorydomain.ErrCategoryNotFound
	}

	subcategory, err := factories.CreateSubcategory(userID, categoryID, input.Name, input.Sequence)
	if err != nil {
		return nil, err
	}

	if err := u.subcategoryRepo.Save(ctx, subcategory); err != nil {
		return nil, err
	}

	return &dtos.SubcategoryOutput{
		ID:         subcategory.ID.String(),
		CategoryID: subcategory.CategoryID.String(),
		Name:       subcategory.Name.String(),
		Sequence:   subcategory.Sequence.Value(),
		CreatedAt:  subcategory.CreatedAt.ValueOr(time.Time{}),
	}, nil
}
