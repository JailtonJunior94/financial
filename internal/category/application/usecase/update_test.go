package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/domain/factories"
	mocks "github.com/jailtonjunior94/financial/internal/category/infrastructure/repositories/mocks"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type UpdateCategoryUseCaseSuite struct {
	suite.Suite

	ctx                context.Context
	obs                observability.Observability
	categoryRepository *mocks.CategoryRepository
}

func TestUpdateCategoryUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UpdateCategoryUseCaseSuite))
}

func (s *UpdateCategoryUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
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
					ParentID: "",
				},
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
			expect: func(output *dtos.CategoryOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal("Transport Updated", output.Name)
				s.Equal(uint(2), output.Sequence)
			},
		},
		{
			name: "deve atualizar categoria com parent_id",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "660e8400-e29b-41d4-a716-446655440001",
				input: &dtos.CategoryInput{
					Name:     "Uber",
					Sequence: 1,
					ParentID: "660e8400-e29b-41d4-a716-446655440002",
				},
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")
					parentID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440002")

					category, _ := factories.CreateCategory(userID.String(), "", "Transport", 1)
					category.ID = categoryID

					s.categoryRepository.
						EXPECT().
						FindByID(s.ctx, userID, categoryID).
						Return(category, nil).
						Once()

					s.categoryRepository.
						EXPECT().
						CheckCycleExists(s.ctx, userID, categoryID, parentID).
						Return(false, nil).
						Once()

					s.categoryRepository.
						EXPECT().
						Update(s.ctx, mock.AnythingOfType("*entities.Category")).
						Return(nil).
						Once()

					return s.categoryRepository
				}(),
			},
			expect: func(output *dtos.CategoryOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal("Uber", output.Name)
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
			name: "deve retornar erro quando categoria for seu próprio pai",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "660e8400-e29b-41d4-a716-446655440001",
				input: &dtos.CategoryInput{
					Name:     "Transport",
					Sequence: 1,
					ParentID: "660e8400-e29b-41d4-a716-446655440001", // mesmo ID
				},
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

					return s.categoryRepository
				}(),
			},
			expect: func(output *dtos.CategoryOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Equal(customErrors.ErrCategoryCycle, err)
			},
		},
		{
			name: "deve retornar erro quando houver ciclo na hierarquia",
			args: args{
				userID:     "550e8400-e29b-41d4-a716-446655440000",
				categoryID: "660e8400-e29b-41d4-a716-446655440001",
				input: &dtos.CategoryInput{
					Name:     "Transport",
					Sequence: 1,
					ParentID: "660e8400-e29b-41d4-a716-446655440002",
				},
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					userID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")
					categoryID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440001")
					parentID, _ := vos.NewUUIDFromString("660e8400-e29b-41d4-a716-446655440002")

					category, _ := factories.CreateCategory(userID.String(), "", "Transport", 1)
					category.ID = categoryID

					s.categoryRepository.
						EXPECT().
						FindByID(s.ctx, userID, categoryID).
						Return(category, nil).
						Once()

					s.categoryRepository.
						EXPECT().
						CheckCycleExists(s.ctx, userID, categoryID, parentID).
						Return(true, nil).
						Once()

					return s.categoryRepository
				}(),
			},
			expect: func(output *dtos.CategoryOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Equal(customErrors.ErrCategoryCycle, err)
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
			uc := NewUpdateCategoryUseCase(s.obs, scenario.dependencies.categoryRepository)
			output, err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.categoryID, scenario.args.input)

			// Assert
			scenario.expect(output, err)
		})
	}
}
