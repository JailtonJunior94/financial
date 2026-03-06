package usecase

import (
"context"
"errors"
"testing"

"github.com/stretchr/testify/mock"
"github.com/stretchr/testify/suite"

"github.com/JailtonJunior94/devkit-go/pkg/observability"
"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
"github.com/JailtonJunior94/devkit-go/pkg/vos"
"github.com/jailtonjunior94/financial/internal/category/application/dtos"
mocks "github.com/jailtonjunior94/financial/internal/category/infrastructure/repositories/mocks"
customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

type UpdateCategoryUseCaseSuite struct {
suite.Suite

ctx                context.Context
obs                observability.Observability
fm                 *metrics.FinancialMetrics
categoryRepository *mocks.CategoryRepository
}

func TestUpdateCategoryUseCaseSuite(t *testing.T) {
suite.Run(t, new(UpdateCategoryUseCaseSuite))
}

func (s *UpdateCategoryUseCaseSuite) SetupTest() {
s.obs = fake.NewProvider()
s.fm = metrics.NewTestFinancialMetrics()
s.ctx = context.Background()
s.categoryRepository = mocks.NewCategoryRepository(s.T())
}

func (s *UpdateCategoryUseCaseSuite) TestExecute() {
type args struct {
userID     string
categoryID string
input      *dtos.CategoryInput
}

type dependencies struct {
categoryRepository *mocks.CategoryRepository
}

scenarios := []struct {
name         string
args         args
dependencies dependencies
expect       func(output *dtos.CategoryOutput, err error)
}{
{
name: "deve atualizar categoria com sucesso",
args: args{
userID:     "550e8400-e29b-41d4-a716-446655440000",
categoryID: "660e8400-e29b-41d4-a716-446655440001",
input: &dtos.CategoryInput{
Name:     "Transport Updated",
Sequence: 2,
},
},
dependencies: dependencies{
categoryRepository: func() *mocks.CategoryRepository {
userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")
category := createCategoryForTest("660e8400-e29b-41d4-a716-446655440001", "Transport", 1)
s.categoryRepository.EXPECT().FindByID(s.ctx, userID, categoryID).Return(category, nil).Once()
s.categoryRepository.EXPECT().Update(s.ctx, mock.AnythingOfType("*entities.Category")).Return(nil).Once()
return s.categoryRepository
}(),
},
expect: func(output *dtos.CategoryOutput, err error) {
s.NoError(err)
s.NotNil(output)
s.Equal("Transport Updated", output.Name)
s.Equal(uint(2), output.Sequence)
},
},
{
name: "deve retornar erro quando categoria não for encontrada",
args: args{
userID:     "550e8400-e29b-41d4-a716-446655440000",
categoryID: "660e8400-e29b-41d4-a716-446655440099",
input: &dtos.CategoryInput{
Name:     "Transport",
Sequence: 1,
},
},
dependencies: dependencies{
categoryRepository: func() *mocks.CategoryRepository {
userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440099")
s.categoryRepository.EXPECT().FindByID(s.ctx, userID, categoryID).Return(nil, nil).Once()
return s.categoryRepository
}(),
},
expect: func(output *dtos.CategoryOutput, err error) {
s.Error(err)
s.Nil(output)
s.Equal(customErrors.ErrCategoryNotFound, err)
},
},
{
name: "deve retornar erro ao falhar ao atualizar no repositório",
args: args{
userID:     "550e8400-e29b-41d4-a716-446655440000",
categoryID: "660e8400-e29b-41d4-a716-446655440001",
input: &dtos.CategoryInput{
Name:     "Transport",
Sequence: 1,
},
},
dependencies: dependencies{
categoryRepository: func() *mocks.CategoryRepository {
userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")
category := createCategoryForTest("660e8400-e29b-41d4-a716-446655440001", "Transport", 1)
s.categoryRepository.EXPECT().FindByID(s.ctx, userID, categoryID).Return(category, nil).Once()
s.categoryRepository.EXPECT().Update(s.ctx, mock.AnythingOfType("*entities.Category")).Return(errors.New("database connection failed")).Once()
return s.categoryRepository
}(),
},
expect: func(output *dtos.CategoryOutput, err error) {
s.Error(err)
s.Nil(output)
s.Contains(err.Error(), "database connection failed")
},
},
}

for _, scenario := range scenarios {
s.Run(scenario.name, func() {
uc := NewUpdateCategoryUseCase(s.obs, s.fm, scenario.dependencies.categoryRepository)
output, err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.categoryID, scenario.args.input)
scenario.expect(output, err)
})
}
}
