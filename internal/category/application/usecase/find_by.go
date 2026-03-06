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
FindCategoryByUseCase interface {
Execute(ctx context.Context, userID, id string) (*dtos.CategoryOutput, error)
}

findCategoryByUseCase struct {
o11y            observability.Observability
fm              *metrics.FinancialMetrics
categoryRepo    interfaces.CategoryRepository
subcategoryRepo interfaces.SubcategoryRepository
}
)

func NewFindCategoryByUseCase(
o11y observability.Observability,
fm *metrics.FinancialMetrics,
categoryRepo interfaces.CategoryRepository,
subcategoryRepo interfaces.SubcategoryRepository,
) FindCategoryByUseCase {
return &findCategoryByUseCase{
o11y:            o11y,
fm:              fm,
categoryRepo:    categoryRepo,
subcategoryRepo: subcategoryRepo,
}
}

func (u *findCategoryByUseCase) Execute(ctx context.Context, userID, id string) (*dtos.CategoryOutput, error) {
ctx, span := u.o11y.Tracer().Start(ctx, "find_category_by_usecase.execute")
defer span.End()

user, err := vos.NewUUIDFromString(userID)
if err != nil {
return nil, err
}

categoryID, err := vos.NewUUIDFromString(id)
if err != nil {
return nil, err
}

category, err := u.categoryRepo.FindByID(ctx, user, categoryID)
if err != nil {
return nil, err
}

if category == nil {
return nil, customErrors.ErrCategoryNotFound
}

subcategories, err := u.subcategoryRepo.FindByCategoryID(ctx, category.ID)
if err != nil {
return nil, err
}

subOutputs := make([]dtos.SubcategoryOutput, 0, len(subcategories))
for _, s := range subcategories {
subOutputs = append(subOutputs, dtos.SubcategoryOutput{
ID:         s.ID.String(),
CategoryID: s.CategoryID.String(),
Name:       s.Name.String(),
Sequence:   s.Sequence.Value(),
CreatedAt:  s.CreatedAt.ValueOr(time.Time{}),
})
}

return &dtos.CategoryOutput{
ID:            category.ID.String(),
Name:          category.Name.String(),
Sequence:      category.Sequence.Value(),
CreatedAt:     category.CreatedAt.ValueOr(time.Time{}),
Subcategories: subOutputs,
}, nil
}
