package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	categorydomain "github.com/jailtonjunior94/financial/internal/category/domain"
	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/internal/category/domain/factories"
	mocks "github.com/jailtonjunior94/financial/internal/category/infrastructure/repositories/mocks"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

type FindSubcategoryByUseCaseSuite struct {
	suite.Suite

	ctx                   context.Context
	obs                   observability.Observability
	fm                    *metrics.FinancialMetrics
	categoryRepository    *mocks.CategoryRepository
	subcategoryRepository *mocks.SubcategoryRepository
}

func TestFindSubcategoryByUseCaseSuite(t *testing.T) {
	suite.Run(t, new(FindSubcategoryByUseCaseSuite))
}

func (s *FindSubcategoryByUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.fm = metrics.NewTestFinancialMetrics()
	s.ctx = context.Background()
	s.categoryRepository = mocks.NewCategoryRepository(s.T())
	s.subcategoryRepository = mocks.NewSubcategoryRepository(s.T())
}

func (s *FindSubcategoryByUseCaseSuite) TestExecute() {
	type args struct {
		userID        string
		categoryID    string
		subcategoryID string
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
			name: "deve encontrar subcategoria com sucesso",
			args: args{
				userID:        "550e8400-e29b-41d4-a716-446655440000",
				categoryID:    "660e8400-e29b-41d4-a716-446655440001",
				subcategoryID: "770e8400-e29b-41d4-a716-446655440001",
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
				s.NoError(err)
				s.NotNil(output)
				s.Equal("Uber", output.Name)
				s.Equal(uint(1), output.Sequence)
				s.Equal("770e8400-e29b-41d4-a716-446655440001", output.ID)
			},
		},
		{
			name: "deve retornar erro quando subcategoria não for encontrada",
			args: args{
				userID:        "550e8400-e29b-41d4-a716-446655440000",
				categoryID:    "660e8400-e29b-41d4-a716-446655440001",
				subcategoryID: "770e8400-e29b-41d4-a716-446655440001",
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
				s.Equal(categorydomain.ErrCategoryNotFound, err)
			},
		},
		{
			name: "deve retornar erro ao falhar ao buscar do repositório",
			args: args{
				userID:        "550e8400-e29b-41d4-a716-446655440000",
				categoryID:    "660e8400-e29b-41d4-a716-446655440001",
				subcategoryID: "770e8400-e29b-41d4-a716-446655440001",
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")
					s.categoryRepository.EXPECT().FindByID(s.ctx, userID, categoryID).Return(nil, errors.New("database connection failed")).Once()
					return s.categoryRepository
				}(),
				subcategoryRepository: s.subcategoryRepository,
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
			uc := NewFindSubcategoryByUseCase(s.obs, s.fm, scenario.dependencies.categoryRepository, scenario.dependencies.subcategoryRepository)
			output, err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.categoryID, scenario.args.subcategoryID)
			scenario.expect(output, err)
		})
	}
}

func makeSubcategoryForTest(id, name string, sequence uint) *entities.Subcategory {
	sub, _ := factories.CreateSubcategory(
		"550e8400-e29b-41d4-a716-446655440000",
		"660e8400-e29b-41d4-a716-446655440001",
		name,
		sequence,
	)
	subID, _ := vos.NewUUIDFromString(id)
	sub.ID = subID
	sub.CreatedAt = vos.NewNullableTime(time.Now())
	return sub
}
