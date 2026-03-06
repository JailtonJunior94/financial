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

type UpdateSubcategoryUseCaseSuite struct {
	suite.Suite

	ctx                   context.Context
	obs                   observability.Observability
	fm                    *metrics.FinancialMetrics
	categoryRepository    *mocks.CategoryRepository
	subcategoryRepository *mocks.SubcategoryRepository
}

func TestUpdateSubcategoryUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UpdateSubcategoryUseCaseSuite))
}

func (s *UpdateSubcategoryUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.fm = metrics.NewTestFinancialMetrics()
	s.ctx = context.Background()
	s.categoryRepository = mocks.NewCategoryRepository(s.T())
	s.subcategoryRepository = mocks.NewSubcategoryRepository(s.T())
}

func (s *UpdateSubcategoryUseCaseSuite) TestExecute() {
	type args struct {
		userID        string
		categoryID    string
		subcategoryID string
		input         *dtos.SubcategoryInput
	}

	type dependencies struct {
		categoryRepository    *mocks.CategoryRepository
		subcategoryRepository *mocks.SubcategoryRepository
	}

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       func(output *dtos.SubcategoryOutput, err error)
	}{
		{
			name: "deve atualizar subcategoria com sucesso",
			args: args{
				userID:        "550e8400-e29b-41d4-a716-446655440000",
				categoryID:    "660e8400-e29b-41d4-a716-446655440001",
				subcategoryID: "770e8400-e29b-41d4-a716-446655440001",
				input: &dtos.SubcategoryInput{
					Name:     "Uber Updated",
					Sequence: 2,
				},
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")
					category := createCategoryForTest("660e8400-e29b-41d4-a716-446655440001", "Transport", 1)
					s.categoryRepository.EXPECT().FindByID(s.ctx, userID, categoryID).Return(category, nil).Once()
					return s.categoryRepository
				}(),
				subcategoryRepository: func() *mocks.SubcategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")
					subID, _ := vos.NewUUIDFromString("770e8400-e29b-41d4-a716-446655440001")
					sub := makeSubcategoryForTest("770e8400-e29b-41d4-a716-446655440001", "Uber", 1)
					s.subcategoryRepository.EXPECT().FindByID(s.ctx, userID, categoryID, subID).Return(sub, nil).Once()
					s.subcategoryRepository.EXPECT().
						Update(s.ctx, mock.AnythingOfType("*entities.Subcategory")).
						Return(nil).
						Once()
					return s.subcategoryRepository
				}(),
			},
			expect: func(output *dtos.SubcategoryOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal("Uber Updated", output.Name)
				s.Equal(uint(2), output.Sequence)
			},
		},
		{
			name: "deve retornar erro quando subcategoria não for encontrada",
			args: args{
				userID:        "550e8400-e29b-41d4-a716-446655440000",
				categoryID:    "660e8400-e29b-41d4-a716-446655440001",
				subcategoryID: "770e8400-e29b-41d4-a716-446655440001",
				input: &dtos.SubcategoryInput{
					Name:     "Uber",
					Sequence: 1,
				},
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")
					category := createCategoryForTest("660e8400-e29b-41d4-a716-446655440001", "Transport", 1)
					s.categoryRepository.EXPECT().FindByID(s.ctx, userID, categoryID).Return(category, nil).Once()
					return s.categoryRepository
				}(),
				subcategoryRepository: func() *mocks.SubcategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")
					subID, _ := vos.NewUUIDFromString("770e8400-e29b-41d4-a716-446655440001")
					s.subcategoryRepository.EXPECT().FindByID(s.ctx, userID, categoryID, subID).Return(nil, nil).Once()
					return s.subcategoryRepository
				}(),
			},
			expect: func(output *dtos.SubcategoryOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Equal(customErrors.ErrSubcategoryNotFound, err)
			},
		},
		{
			name: "deve retornar erro quando categoria não for encontrada",
			args: args{
				userID:        "550e8400-e29b-41d4-a716-446655440000",
				categoryID:    "660e8400-e29b-41d4-a716-446655440001",
				subcategoryID: "770e8400-e29b-41d4-a716-446655440001",
				input: &dtos.SubcategoryInput{
					Name:     "Uber",
					Sequence: 1,
				},
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")
					s.categoryRepository.EXPECT().FindByID(s.ctx, userID, categoryID).Return(nil, nil).Once()
					return s.categoryRepository
				}(),
				subcategoryRepository: s.subcategoryRepository,
			},
			expect: func(output *dtos.SubcategoryOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Equal(customErrors.ErrCategoryNotFound, err)
			},
		},
		{
			name: "deve retornar erro ao atualizar subcategoria com nome vazio",
			args: args{
				userID:        "550e8400-e29b-41d4-a716-446655440000",
				categoryID:    "660e8400-e29b-41d4-a716-446655440001",
				subcategoryID: "770e8400-e29b-41d4-a716-446655440001",
				input: &dtos.SubcategoryInput{
					Name:     "",
					Sequence: 1,
				},
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")
					category := createCategoryForTest("660e8400-e29b-41d4-a716-446655440001", "Transport", 1)
					s.categoryRepository.EXPECT().FindByID(s.ctx, userID, categoryID).Return(category, nil).Once()
					return s.categoryRepository
				}(),
				subcategoryRepository: func() *mocks.SubcategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")
					subID, _ := vos.NewUUIDFromString("770e8400-e29b-41d4-a716-446655440001")
					sub := makeSubcategoryForTest("770e8400-e29b-41d4-a716-446655440001", "Uber", 1)
					s.subcategoryRepository.EXPECT().FindByID(s.ctx, userID, categoryID, subID).Return(sub, nil).Once()
					return s.subcategoryRepository
				}(),
			},
			expect: func(output *dtos.SubcategoryOutput, err error) {
				s.Error(err)
				s.Nil(output)
			},
		},
		{
			name: "deve retornar erro ao falhar ao atualizar no repositório",
			args: args{
				userID:        "550e8400-e29b-41d4-a716-446655440000",
				categoryID:    "660e8400-e29b-41d4-a716-446655440001",
				subcategoryID: "770e8400-e29b-41d4-a716-446655440001",
				input: &dtos.SubcategoryInput{
					Name:     "Uber Updated",
					Sequence: 2,
				},
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")
					category := createCategoryForTest("660e8400-e29b-41d4-a716-446655440001", "Transport", 1)
					s.categoryRepository.EXPECT().FindByID(s.ctx, userID, categoryID).Return(category, nil).Once()
					return s.categoryRepository
				}(),
				subcategoryRepository: func() *mocks.SubcategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")
					subID, _ := vos.NewUUIDFromString("770e8400-e29b-41d4-a716-446655440001")
					sub := makeSubcategoryForTest("770e8400-e29b-41d4-a716-446655440001", "Uber", 1)
					s.subcategoryRepository.EXPECT().FindByID(s.ctx, userID, categoryID, subID).Return(sub, nil).Once()
					s.subcategoryRepository.EXPECT().
						Update(s.ctx, mock.AnythingOfType("*entities.Subcategory")).
						Return(errors.New("database connection failed")).
						Once()
					return s.subcategoryRepository
				}(),
			},
			expect: func(output *dtos.SubcategoryOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "database connection failed")
			},
		},
{
name: "deve retornar erro com userID inválido",
args: args{
userID:        "invalid-uuid",
categoryID:    "660e8400-e29b-41d4-a716-446655440001",
subcategoryID: "770e8400-e29b-41d4-a716-446655440001",
input:         &dtos.SubcategoryInput{Name: "Uber", Sequence: 1},
},
dependencies: dependencies{
categoryRepository:    s.categoryRepository,
subcategoryRepository: s.subcategoryRepository,
},
expect: func(output *dtos.SubcategoryOutput, err error) {
s.Error(err)
s.Nil(output)
},
},
{
name: "deve retornar erro com categoryID inválido",
args: args{
userID:        "550e8400-e29b-41d4-a716-446655440000",
categoryID:    "invalid-uuid",
subcategoryID: "770e8400-e29b-41d4-a716-446655440001",
input:         &dtos.SubcategoryInput{Name: "Uber", Sequence: 1},
},
dependencies: dependencies{
categoryRepository:    s.categoryRepository,
subcategoryRepository: s.subcategoryRepository,
},
expect: func(output *dtos.SubcategoryOutput, err error) {
s.Error(err)
s.Nil(output)
},
},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			uc := NewUpdateSubcategoryUseCase(s.obs, s.fm, scenario.dependencies.categoryRepository, scenario.dependencies.subcategoryRepository)
			output, err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.categoryID, scenario.args.subcategoryID, scenario.args.input)
			scenario.expect(output, err)
		})
	}
}
