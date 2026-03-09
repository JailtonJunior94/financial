package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/domain/factories"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

type (
	CreateCategoryUseCase interface {
		Execute(ctx context.Context, userID string, input *dtos.CategoryInput) (*dtos.CategoryOutput, error)
	}

	createCategoryUseCase struct {
		o11y       observability.Observability
		fm         *metrics.FinancialMetrics
		repository interfaces.CategoryRepository
	}
)

func NewCreateCategoryUseCase(
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
	repository interfaces.CategoryRepository,
) CreateCategoryUseCase {
	return &createCategoryUseCase{
		o11y:       o11y,
		fm:         fm,
		repository: repository,
	}
}

func (u *createCategoryUseCase) Execute(ctx context.Context, userID string, input *dtos.CategoryInput) (*dtos.CategoryOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "create_category_usecase.execute")
	defer span.End()

	category, err := factories.CreateCategory(userID, input.Name, input.Sequence)
	if err != nil {
		return nil, err
	}

	if err := u.repository.Save(ctx, category); err != nil {
		return nil, err
	}

	return &dtos.CategoryOutput{
		ID:        category.ID.String(),
		Name:      category.Name.String(),
		Sequence:  category.Sequence.Value(),
		CreatedAt: category.CreatedAt.ValueOr(time.Time{}),
	}, nil
}
