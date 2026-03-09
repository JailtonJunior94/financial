package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	mocks "github.com/jailtonjunior94/financial/internal/category/infrastructure/repositories/mocks"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

type FindCategoryPaginatedUseCaseSuite struct {
	suite.Suite

	ctx                context.Context
	obs                observability.Observability
	fm                 *metrics.FinancialMetrics
	categoryRepository *mocks.CategoryRepository
}

func TestFindCategoryPaginatedUseCaseSuite(t *testing.T) {
	suite.Run(t, new(FindCategoryPaginatedUseCaseSuite))
}

func (s *FindCategoryPaginatedUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.fm = metrics.NewTestFinancialMetrics()
	s.ctx = context.Background()
	s.categoryRepository = mocks.NewCategoryRepository(s.T())
}

func (s *FindCategoryPaginatedUseCaseSuite) TestExecute() {
	type args struct {
		input FindCategoryPaginatedInput
	}

	type dependencies struct {
		categoryRepository *mocks.CategoryRepository
	}

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       func(output *FindCategoryPaginatedOutput, err error)
	}{
		{
			name: "deve listar categorias sem cursor",
			args: args{
				input: FindCategoryPaginatedInput{
					UserID: "550e8400-e29b-41d4-a716-446655440000",
					Limit:  2,
					Cursor: "",
				},
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					categories := []*entities.Category{
						createCategoryForTest("660e8400-e29b-41d4-a716-446655440001", "Transport", 1),
						createCategoryForTest("660e8400-e29b-41d4-a716-446655440002", "Food", 2),
					}
					s.categoryRepository.EXPECT().
						ListPaginated(s.ctx, mock.AnythingOfType("interfaces.ListCategoriesParams")).
						Return(categories, nil).
						Once()
					return s.categoryRepository
				}(),
			},
			expect: func(output *FindCategoryPaginatedOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Categories, 2)
				s.Nil(output.NextCursor)
			},
		},
		{
			name: "deve retornar lista com hasNext quando há mais itens",
			args: args{
				input: FindCategoryPaginatedInput{
					UserID: "550e8400-e29b-41d4-a716-446655440000",
					Limit:  2,
					Cursor: "",
				},
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					// limit+1 = 3 items returned means hasNext = true
					categories := []*entities.Category{
						createCategoryForTest("660e8400-e29b-41d4-a716-446655440001", "Transport", 1),
						createCategoryForTest("660e8400-e29b-41d4-a716-446655440002", "Food", 2),
						createCategoryForTest("660e8400-e29b-41d4-a716-446655440003", "Entertainment", 3),
					}
					s.categoryRepository.EXPECT().
						ListPaginated(s.ctx, mock.AnythingOfType("interfaces.ListCategoriesParams")).
						Return(categories, nil).
						Once()
					return s.categoryRepository
				}(),
			},
			expect: func(output *FindCategoryPaginatedOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Categories, 2)
				s.NotNil(output.NextCursor)
			},
		},
		{
			name: "deve retornar lista vazia",
			args: args{
				input: FindCategoryPaginatedInput{
					UserID: "550e8400-e29b-41d4-a716-446655440000",
					Limit:  20,
					Cursor: "",
				},
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					s.categoryRepository.EXPECT().
						ListPaginated(s.ctx, mock.AnythingOfType("interfaces.ListCategoriesParams")).
						Return([]*entities.Category{}, nil).
						Once()
					return s.categoryRepository
				}(),
			},
			expect: func(output *FindCategoryPaginatedOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Categories, 0)
				s.Nil(output.NextCursor)
			},
		},
		{
			name: "deve retornar erro ao falhar ao buscar do repositório",
			args: args{
				input: FindCategoryPaginatedInput{
					UserID: "550e8400-e29b-41d4-a716-446655440000",
					Limit:  20,
					Cursor: "",
				},
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					s.categoryRepository.EXPECT().
						ListPaginated(s.ctx, mock.AnythingOfType("interfaces.ListCategoriesParams")).
						Return(nil, errors.New("db error")).
						Once()
					return s.categoryRepository
				}(),
			},
			expect: func(output *FindCategoryPaginatedOutput, err error) {
				s.Error(err)
				s.Nil(output)
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			uc := NewFindCategoryPaginatedUseCase(s.obs, s.fm, scenario.dependencies.categoryRepository)
			output, err := uc.Execute(s.ctx, scenario.args.input)
			scenario.expect(output, err)
		})
	}
}
