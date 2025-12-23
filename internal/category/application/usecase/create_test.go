package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	mocks "github.com/jailtonjunior94/financial/internal/category/infrastructure/repositories/mocks"
)

type CreateCategoryUseCaseSuite struct {
	suite.Suite

	ctx                context.Context
	categoryRepository *mocks.CategoryRepository
}

func TestCreateCategoryUseCaseSuite(t *testing.T) {
	suite.Run(t, new(CreateCategoryUseCaseSuite))
}

func (s *CreateCategoryUseCaseSuite) SetupTest() {
	s.ctx = context.Background()
	s.categoryRepository = mocks.NewCategoryRepository(s.T())
}

func (s *CreateCategoryUseCaseSuite) TestExecute() {
	type args struct {
		userID string
		input  *dtos.CategoryInput
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
			name: "deve criar categoria com sucesso",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.CategoryInput{
					Name:     "Transport",
					Sequence: 1,
					ParentID: "",
				},
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					s.categoryRepository.
						EXPECT().
						Save(s.ctx, mock.AnythingOfType("*entities.Category")).
						Return(nil).
						Once()
					return s.categoryRepository
				}(),
			},
			expect: func(output *dtos.CategoryOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal("Transport", output.Name)
				s.Equal(uint(1), output.Sequence)
				s.NotEmpty(output.ID)
			},
		},
		{
			name: "deve criar subcategoria com parent_id válido",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.CategoryInput{
					Name:     "Uber",
					Sequence: 1,
					ParentID: "660e8400-e29b-41d4-a716-446655440000",
				},
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					s.categoryRepository.
						EXPECT().
						Save(s.ctx, mock.AnythingOfType("*entities.Category")).
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
			name: "deve retornar erro ao criar categoria com nome vazio",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.CategoryInput{
					Name:     "",
					Sequence: 1,
				},
			},
			dependencies: dependencies{
				categoryRepository: s.categoryRepository,
			},
			expect: func(output *dtos.CategoryOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "invalid category name")
			},
		},
		{
			name: "deve retornar erro ao criar categoria com sequence 0",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.CategoryInput{
					Name:     "Transport",
					Sequence: 0,
				},
			},
			dependencies: dependencies{
				categoryRepository: s.categoryRepository,
			},
			expect: func(output *dtos.CategoryOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "invalid category sequence")
			},
		},
		{
			name: "deve retornar erro ao falhar ao salvar no repositório",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.CategoryInput{
					Name:     "Transport",
					Sequence: 1,
				},
			},
			dependencies: dependencies{
				categoryRepository: func() *mocks.CategoryRepository {
					s.categoryRepository.
						EXPECT().
						Save(s.ctx, mock.AnythingOfType("*entities.Category")).
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
		{
			name: "deve retornar erro com user_id inválido",
			args: args{
				userID: "invalid-uuid",
				input: &dtos.CategoryInput{
					Name:     "Transport",
					Sequence: 1,
				},
			},
			dependencies: dependencies{
				categoryRepository: s.categoryRepository,
			},
			expect: func(output *dtos.CategoryOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "invalid user_id")
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Act
			telemetry := newMockTelemetry()
			uc := NewCreateCategoryUseCase(telemetry, scenario.dependencies.categoryRepository)
			output, err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.input)

			// Assert
			scenario.expect(output, err)
		})
	}
}
