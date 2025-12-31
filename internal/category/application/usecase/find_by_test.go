package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/domain/entities"
	"github.com/jailtonjunior94/financial/internal/category/domain/factories"
	mocks "github.com/jailtonjunior94/financial/internal/category/infrastructure/repositories/mocks"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type FindCategoryByUseCaseSuite struct {
	suite.Suite

	ctx                context.Context
	obs                observability.Observability
	categoryRepository *mocks.CategoryRepository
}

func TestFindCategoryByUseCaseSuite(t *testing.T) {
	suite.Run(t, new(FindCategoryByUseCaseSuite))
}

func (s *FindCategoryByUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.categoryRepository = mocks.NewCategoryRepository(s.T())
}

func (s *FindCategoryByUseCaseSuite) TestExecute() {
	type args struct {
		userID     string
		categoryID string
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
			name: "deve encontrar categoria por ID com sucesso",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "660e8400-e29b-41d4-a716-446655440001",
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")

					category := createCategoryForTest("660e8400-e29b-41d4-a716-446655440001", "Transport", 1)

					s.categoryRepository.
						EXPECT().
						FindByID(s.ctx, userID, categoryID).
						Return(category, nil).
						Once()
					return s.categoryRepository
				}(),
			},
			expect: func(output *dtos.CategoryOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal("Transport", output.Name)
				s.Equal(uint(1), output.Sequence)
				s.Equal("660e8400-e29b-41d4-a716-446655440001", output.ID)
				s.Empty(output.Children)
			},
		},
		{
			name: "deve encontrar categoria com children",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "660e8400-e29b-41d4-a716-446655440001",
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")

					category := createCategoryForTest("660e8400-e29b-41d4-a716-446655440001", "Transport", 1)

					// Adicionar children
					child1 := createCategoryForTest("660e8400-e29b-41d4-a716-446655440002", "Uber", 1)
					child2 := createCategoryForTest("660e8400-e29b-41d4-a716-446655440003", "Bus", 2)
					category.AddChildrens([]entities.Category{*child1, *child2})

					s.categoryRepository.
						EXPECT().
						FindByID(s.ctx, userID, categoryID).
						Return(category, nil).
						Once()
					return s.categoryRepository
				}(),
			},
			expect: func(output *dtos.CategoryOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal("Transport", output.Name)
				s.Len(output.Children, 2)
				s.Equal("Uber", output.Children[0].Name)
				s.Equal("Bus", output.Children[1].Name)
			},
		},
		{
			name: "deve retornar erro quando categoria não for encontrada",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "660e8400-e29b-41d4-a716-446655440099",
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440099")

					s.categoryRepository.
						EXPECT().
						FindByID(s.ctx, userID, categoryID).
						Return(nil, nil).
						Once()
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
			name: "deve retornar erro com user_id inválido",
			args: args{
				userID:     "invalid-uuid",
				categoryID: "660e8400-e29b-41d4-a716-446655440001",
			},
			dependencies: dependencies{
				categoryRepository: s.categoryRepository,
			},
			expect: func(output *dtos.CategoryOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "invalid UUID")
			},
		},
		{
			name: "deve retornar erro com category_id inválido",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "invalid-uuid",
			},
			dependencies: dependencies{
				categoryRepository: s.categoryRepository,
			},
			expect: func(output *dtos.CategoryOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "invalid UUID")
			},
		},
		{
			name: "deve retornar erro ao falhar ao buscar do repositório",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "660e8400-e29b-41d4-a716-446655440001",
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")

					s.categoryRepository.
						EXPECT().
						FindByID(s.ctx, userID, categoryID).
						Return(nil, errors.New("database connection failed")).
						Once()
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
			// Act
			uc := NewFindCategoryByUseCase(s.obs, scenario.dependencies.categoryRepository)
			output, err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.categoryID)

			// Assert
			scenario.expect(output, err)
		})
	}
}

// createCategoryForTest é um helper para criar categorias de teste
func createCategoryForTest(id, name string, sequence uint) *entities.Category {
	category, _ := factories.CreateCategory(
		"550e8400-e29b-41d4-a716-446655440000",
		"",
		name,
		sequence,
	)

	// Sobrescrever o ID gerado automaticamente com o ID de teste
	categoryID, _ := vos.NewUUIDFromString(id)
	category.ID = categoryID
	category.CreatedAt = vos.NewNullableTime(time.Now())

	return category
}
