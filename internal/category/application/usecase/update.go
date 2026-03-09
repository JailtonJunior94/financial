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

type (
	UpdateCategoryUseCase interface {
		Execute(ctx context.Context, userID, id string, input *dtos.CategoryInput) (*dtos.CategoryOutput, error)
	}

	updateCategoryUseCase struct {
		o11y       observability.Observability
		fm         *metrics.FinancialMetrics
		repository interfaces.CategoryRepository
	}
)

func NewUpdateCategoryUseCase(
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
	repository interfaces.CategoryRepository,
) UpdateCategoryUseCase {
	return &updateCategoryUseCase{
		o11y:       o11y,
		fm:         fm,
		repository: repository,
	}
}

func (u *updateCategoryUseCase) Execute(ctx context.Context, userID, id string, input *dtos.CategoryInput) (*dtos.CategoryOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "update_category_usecase.execute")
	defer span.End()

	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return nil, err
	}

	categoryID, err := vos.NewUUIDFromString(id)
	if err != nil {
		return nil, err
	}

	category, err := u.repository.FindByID(ctx, user, categoryID)
	if err != nil {
		return nil, err
	}

	if category == nil {
		return nil, customErrors.ErrCategoryNotFound
	}

	if err := category.Update(input.Name, input.Sequence); err != nil {
		return nil, err
	}

	if err := u.repository.Update(ctx, category); err != nil {
		return nil, err
	}

	return &dtos.CategoryOutput{
		ID:        category.ID.String(),
		Name:      category.Name.String(),
		Sequence:  category.Sequence.Value(),
		CreatedAt: category.CreatedAt.ValueOr(time.Time{}),
	}, nil
}
