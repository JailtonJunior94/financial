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
	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	mocks "github.com/jailtonjunior94/financial/internal/category/infrastructure/repositories/mocks"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

type FindSubcategoriesPaginatedUseCaseSuite struct {
	suite.Suite

	ctx                   context.Context
	obs                   observability.Observability
	fm                    *metrics.FinancialMetrics
	categoryRepository    *mocks.CategoryRepository
	subcategoryRepository *mocks.SubcategoryRepository
}

func TestFindSubcategoriesPaginatedUseCaseSuite(t *testing.T) {
	suite.Run(t, new(FindSubcategoriesPaginatedUseCaseSuite))
}

func (s *FindSubcategoriesPaginatedUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.fm = metrics.NewTestFinancialMetrics()
	s.ctx = context.Background()
	s.categoryRepository = mocks.NewCategoryRepository(s.T())
	s.subcategoryRepository = mocks.NewSubcategoryRepository(s.T())
}

func (s *FindSubcategoriesPaginatedUseCaseSuite) TestExecute() {
	type args struct {
		userID     string
		categoryID string
		limit      int
		cursor     string
	}

	type dependencies struct {
		categoryRepository    *mocks.CategoryRepository
		subcategoryRepository *mocks.SubcategoryRepository
	}

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       func(output *dtos.SubcategoryPaginatedOutput, err error)
	}{
		{
			name: "deve listar subcategorias sem cursor",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "660e8400-e29b-41d4-a716-446655440001",
				limit:      2,
				cursor:     "",
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
					subcategories := []*entities.Subcategory{
						makeSubcategoryForTest("770e8400-e29b-41d4-a716-446655440001", "Uber", 1),
						makeSubcategoryForTest("770e8400-e29b-41d4-a716-446655440002", "Bus", 2),
					}
					s.subcategoryRepository.EXPECT().
						ListPaginated(s.ctx, mock.AnythingOfType("interfaces.ListSubcategoriesParams")).
						Return(subcategories, nil).
						Once()
					return s.subcategoryRepository
				}(),
			},
			expect: func(output *dtos.SubcategoryPaginatedOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Data, 2)
				s.False(output.Pagination.HasNext)
				s.Nil(output.Pagination.NextCursor)
			},
		},
		{
			name: "deve retornar hasNext quando há mais itens",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "660e8400-e29b-41d4-a716-446655440001",
				limit:      2,
				cursor:     "",
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
					// limit+1 = 3 items means hasNext = true
					subcategories := []*entities.Subcategory{
						makeSubcategoryForTest("770e8400-e29b-41d4-a716-446655440001", "Uber", 1),
						makeSubcategoryForTest("770e8400-e29b-41d4-a716-446655440002", "Bus", 2),
						makeSubcategoryForTest("770e8400-e29b-41d4-a716-446655440003", "Taxi", 3),
					}
					s.subcategoryRepository.EXPECT().
						ListPaginated(s.ctx, mock.AnythingOfType("interfaces.ListSubcategoriesParams")).
						Return(subcategories, nil).
						Once()
					return s.subcategoryRepository
				}(),
			},
			expect: func(output *dtos.SubcategoryPaginatedOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Data, 2)
				s.True(output.Pagination.HasNext)
				s.NotNil(output.Pagination.NextCursor)
			},
		},
		{
			name: "deve retornar lista vazia",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "660e8400-e29b-41d4-a716-446655440001",
				limit:      20,
				cursor:     "",
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
					s.subcategoryRepository.EXPECT().
						ListPaginated(s.ctx, mock.AnythingOfType("interfaces.ListSubcategoriesParams")).
						Return([]*entities.Subcategory{}, nil).
						Once()
					return s.subcategoryRepository
				}(),
			},
			expect: func(output *dtos.SubcategoryPaginatedOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Data, 0)
				s.False(output.Pagination.HasNext)
			},
		},
		{
			name: "deve retornar erro quando categoria não for encontrada",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "660e8400-e29b-41d4-a716-446655440001",
				limit:      20,
				cursor:     "",
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
			expect: func(output *dtos.SubcategoryPaginatedOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Equal(customErrors.ErrCategoryNotFound, err)
			},
		},
		{
			name: "deve retornar erro ao falhar ao buscar do repositório",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "660e8400-e29b-41d4-a716-446655440001",
				limit:      20,
				cursor:     "",
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
					s.subcategoryRepository.EXPECT().
						ListPaginated(s.ctx, mock.AnythingOfType("interfaces.ListSubcategoriesParams")).
						Return(nil, errors.New("database connection failed")).
						Once()
					return s.subcategoryRepository
				}(),
			},
			expect: func(output *dtos.SubcategoryPaginatedOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "database connection failed")
			},
		},
		{
			name: "deve retornar erro com userID inválido",
			args: args{
				userID:     "invalid-uuid",
				categoryID: "660e8400-e29b-41d4-a716-446655440001",
				limit:      20,
				cursor:     "",
			},
			dependencies: dependencies{
				categoryRepository:    s.categoryRepository,
				subcategoryRepository: s.subcategoryRepository,
			},
			expect: func(output *dtos.SubcategoryPaginatedOutput, err error) {
				s.Error(err)
				s.Nil(output)
			},
		},
		{
			name: "deve retornar erro com categoryID inválido",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "invalid-uuid",
				limit:      20,
				cursor:     "",
			},
			dependencies: dependencies{
				categoryRepository:    s.categoryRepository,
				subcategoryRepository: s.subcategoryRepository,
			},
			expect: func(output *dtos.SubcategoryPaginatedOutput, err error) {
				s.Error(err)
				s.Nil(output)
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			uc := NewFindSubcategoriesPaginatedUseCase(s.obs, s.fm, scenario.dependencies.categoryRepository, scenario.dependencies.subcategoryRepository)
			output, err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.categoryID, scenario.args.limit, scenario.args.cursor)
			scenario.expect(output, err)
		})
	}
}
