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

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type FindCategoryUseCaseSuite struct {
	suite.Suite

	ctx                context.Context
	obs                observability.Observability
	categoryRepository *mocks.CategoryRepository
}

func TestFindCategoryUseCaseSuite(t *testing.T) {
	suite.Run(t, new(FindCategoryUseCaseSuite))
}

func (s *FindCategoryUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.categoryRepository = mocks.NewCategoryRepository(s.T())
}

func (s *FindCategoryUseCaseSuite) TestExecute() {
	type args struct {
		userID string
	}

	type dependencies struct {
		categoryRepository *mocks.CategoryRepository
	}

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       func(output []*dtos.CategoryOutput, err error)
	}{
		{
			name: "deve listar categorias com sucesso",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")

					categories := []*entities.Category{
						createTestCategory("660e8400-e29b-41d4-a716-446655440001", "Transport", 1),
						createTestCategory("660e8400-e29b-41d4-a716-446655440002", "Food", 2),
						createTestCategory("660e8400-e29b-41d4-a716-446655440003", "Entertainment", 3),
					}

					s.categoryRepository.
						EXPECT().
						List(s.ctx, userID).
						Return(categories, nil).
						Once()
					return s.categoryRepository
				}(),
			},
			expect: func(output []*dtos.CategoryOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output, 3)
				s.Equal("Transport", output[0].Name)
				s.Equal("Food", output[1].Name)
				s.Equal("Entertainment", output[2].Name)
				s.Equal(uint(1), output[0].Sequence)
				s.Equal(uint(2), output[1].Sequence)
				s.Equal(uint(3), output[2].Sequence)
			},
		},
		{
			name: "deve retornar lista vazia quando não houver categorias",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")

					s.categoryRepository.
						EXPECT().
						List(s.ctx, userID).
						Return([]*entities.Category{}, nil).
						Once()
					return s.categoryRepository
				}(),
			},
			expect: func(output []*dtos.CategoryOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Empty(output)
			},
		},
		{
			name: "deve retornar erro com user_id inválido",
			args: args{
				userID: "invalid-uuid",
			},
			dependencies: dependencies{
				categoryRepository: s.categoryRepository,
			},
			expect: func(output []*dtos.CategoryOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "invalid UUID")
			},
		},
		{
			name: "deve retornar erro ao falhar ao listar do repositório",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")

					s.categoryRepository.
						EXPECT().
						List(s.ctx, userID).
						Return(nil, errors.New("database connection failed")).
						Once()
					return s.categoryRepository
				}(),
			},
			expect: func(output []*dtos.CategoryOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "database connection failed")
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Act
			uc := NewFindCategoryUseCase(s.obs, scenario.dependencies.categoryRepository)
			output, err := uc.Execute(s.ctx, scenario.args.userID)

			// Assert
			scenario.expect(output, err)
		})
	}
}

// createTestCategory é um helper para criar categorias de teste
func createTestCategory(id, name string, sequence uint) *entities.Category {
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
