package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financial/internal/category/domain/factories"
	mocks "github.com/jailtonjunior94/financial/internal/category/infrastructure/repositories/mocks"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type RemoveCategoryUseCaseSuite struct {
	suite.Suite

	ctx                context.Context
	obs                observability.Observability
	categoryRepository *mocks.CategoryRepository
}

func TestRemoveCategoryUseCaseSuite(t *testing.T) {
	suite.Run(t, new(RemoveCategoryUseCaseSuite))
}

func (s *RemoveCategoryUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.categoryRepository = mocks.NewCategoryRepository(s.T())
}

func (s *RemoveCategoryUseCaseSuite) TestExecute() {
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
		expect       func(err error)
	}{
		{
			name: "deve remover categoria com sucesso",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "660e8400-e29b-41d4-a716-446655440001",
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")

					category, _ := factories.CreateCategory(userID.String(), "", "Transport", 1)
					category.ID = categoryID

					s.categoryRepository.
						EXPECT().
						FindByID(s.ctx, userID, categoryID).
						Return(category, nil).
						Once()

					s.categoryRepository.
						EXPECT().
						Update(s.ctx, mock.AnythingOfType("*entities.Category")).
						Return(nil).
						Once()

					return s.categoryRepository
				}(),
			},
			expect: func(err error) {
				s.NoError(err)
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
			expect: func(err error) {
				s.Error(err)
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
			expect: func(err error) {
				s.Error(err)
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
			expect: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "invalid UUID")
			},
		},
		{
			name: "deve retornar erro ao falhar ao atualizar no repositório",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "660e8400-e29b-41d4-a716-446655440001",
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")

					category, _ := factories.CreateCategory(userID.String(), "", "Transport", 1)
					category.ID = categoryID

					s.categoryRepository.
						EXPECT().
						FindByID(s.ctx, userID, categoryID).
						Return(category, nil).
						Once()

					s.categoryRepository.
						EXPECT().
						Update(s.ctx, mock.AnythingOfType("*entities.Category")).
						Return(errors.New("database connection failed")).
						Once()

					return s.categoryRepository
				}(),
			},
			expect: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "database connection failed")
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Act
			uc := NewRemoveCategoryUseCase(s.obs, scenario.dependencies.categoryRepository)
			err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.categoryID)

			// Assert
			scenario.expect(err)
		})
	}
}
