package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/jailtonjunior94/financial/internal/card/application/dtos"
	repositoryMock "github.com/jailtonjunior94/financial/internal/card/infrastructure/repositories/mocks"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
)

type CreateCardUseCaseSuite struct {
	suite.Suite

	ctx            context.Context
	obs            observability.Observability
	cardRepository *repositoryMock.CardRepository
}

func TestCreateCardUseCaseSuite(t *testing.T) {
	suite.Run(t, new(CreateCardUseCaseSuite))
}

func (s *CreateCardUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.cardRepository = repositoryMock.NewCardRepository(s.T())
}

func (s *CreateCardUseCaseSuite) TestExecute() {
	type args struct {
		userID string
		input  *dtos.CardInput
	}

	type dependencies struct {
		cardRepository *repositoryMock.CardRepository
	}

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       func(output *dtos.CardOutput, err error)
	}{
		{
			name: "deve criar cartão com sucesso",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.CardInput{
					Name:   "Nubank",
					DueDay: 15,
				},
			},
			dependencies: dependencies{
				cardRepository: func() *repositoryMock.CardRepository {
					s.cardRepository.
						EXPECT().
						Save(s.ctx, mock.AnythingOfType("*entities.Card")).
						Return(nil).
						Once()
					return s.cardRepository
				}(),
			},
			expect: func(output *dtos.CardOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal("Nubank", output.Name)
				s.Equal(15, output.DueDay)
				s.NotEmpty(output.ID)
			},
		},
		{
			name: "deve criar cartão com due_day 1",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.CardInput{
					Name:   "Inter",
					DueDay: 1,
				},
			},
			dependencies: dependencies{
				cardRepository: func() *repositoryMock.CardRepository {
					s.cardRepository.
						EXPECT().
						Save(s.ctx, mock.AnythingOfType("*entities.Card")).
						Return(nil).
						Once()
					return s.cardRepository
				}(),
			},
			expect: func(output *dtos.CardOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal("Inter", output.Name)
				s.Equal(1, output.DueDay)
			},
		},
		{
			name: "deve criar cartão com due_day 31",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.CardInput{
					Name:   "Banco do Brasil",
					DueDay: 31,
				},
			},
			dependencies: dependencies{
				cardRepository: func() *repositoryMock.CardRepository {
					s.cardRepository.
						EXPECT().
						Save(s.ctx, mock.AnythingOfType("*entities.Card")).
						Return(nil).
						Once()
					return s.cardRepository
				}(),
			},
			expect: func(output *dtos.CardOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal("Banco do Brasil", output.Name)
				s.Equal(31, output.DueDay)
			},
		},
		{
			name: "deve retornar erro ao criar cartão com nome vazio",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.CardInput{
					Name:   "",
					DueDay: 15,
				},
			},
			dependencies: dependencies{
				cardRepository: s.cardRepository,
			},
			expect: func(output *dtos.CardOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "invalid card name")
			},
		},
		{
			name: "deve retornar erro ao criar cartão com due_day 0",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.CardInput{
					Name:   "Nubank",
					DueDay: 0,
				},
			},
			dependencies: dependencies{
				cardRepository: s.cardRepository,
			},
			expect: func(output *dtos.CardOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "invalid due day")
			},
		},
		{
			name: "deve retornar erro ao criar cartão com due_day 32",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.CardInput{
					Name:   "Nubank",
					DueDay: 32,
				},
			},
			dependencies: dependencies{
				cardRepository: s.cardRepository,
			},
			expect: func(output *dtos.CardOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "invalid due day")
			},
		},
		{
			name: "deve retornar erro ao falhar ao salvar no repositório",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.CardInput{
					Name:   "Nubank",
					DueDay: 15,
				},
			},
			dependencies: dependencies{
				cardRepository: func() *repositoryMock.CardRepository {
					s.cardRepository.
						EXPECT().
						Save(s.ctx, mock.AnythingOfType("*entities.Card")).
						Return(errors.New("database connection failed")).
						Once()
					return s.cardRepository
				}(),
			},
			expect: func(output *dtos.CardOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "database connection failed")
			},
		},
		{
			name: "deve retornar erro com user_id inválido",
			args: args{
				userID: "invalid-uuid",
				input: &dtos.CardInput{
					Name:   "Nubank",
					DueDay: 15,
				},
			},
			dependencies: dependencies{
				cardRepository: s.cardRepository,
			},
			expect: func(output *dtos.CardOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "invalid user_id")
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Act
			uc := NewCreateCardUseCase(s.obs, scenario.dependencies.cardRepository)
			output, err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.input)

			// Assert
			scenario.expect(output, err)
		})
	}
}
