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
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

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
			name: "deve criar cartão de crédito com sucesso",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.CardInput{
					Name:           "Nubank",
					Type:           "credit",
					Flag:           "mastercard",
					LastFourDigits: "1234",
					DueDay:         intPtrUC(15),
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
				s.Equal("credit", output.Type)
				s.Equal("mastercard", output.Flag)
				s.Equal("1234", output.LastFourDigits)
				s.NotNil(output.DueDay)
				s.Equal(15, *output.DueDay)
				s.NotEmpty(output.ID)
			},
		},
		{
			name: "deve criar cartão de débito sem due_day",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.CardInput{
					Name:           "Nubank Debito",
					Type:           "debit",
					Flag:           "visa",
					LastFourDigits: "5678",
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
				s.Equal("debit", output.Type)
				s.Nil(output.DueDay)
				s.Nil(output.ClosingOffsetDays)
			},
		},
		{
			name: "deve criar cartão de crédito com closing_offset_days default",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.CardInput{
					Name:           "Inter",
					Type:           "credit",
					Flag:           "elo",
					LastFourDigits: "9999",
					DueDay:         intPtrUC(10),
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
				s.NotNil(output.ClosingOffsetDays)
				s.Equal(7, *output.ClosingOffsetDays)
			},
		},
		{
			name: "deve retornar erro com tipo inválido",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.CardInput{
					Name:           "Nubank",
					Type:           "prepaid",
					Flag:           "visa",
					LastFourDigits: "1234",
				},
			},
			dependencies: dependencies{
				cardRepository: s.cardRepository,
			},
			expect: func(output *dtos.CardOutput, err error) {
				s.Error(err)
				s.Nil(output)
			},
		},
		{
			name: "deve retornar erro ao falhar ao salvar no repositório",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.CardInput{
					Name:           "Nubank",
					Type:           "credit",
					Flag:           "visa",
					LastFourDigits: "1234",
					DueDay:         intPtrUC(15),
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
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			cardMetrics := metrics.NewTestCardMetrics()
			uc := NewCreateCardUseCase(s.obs, scenario.dependencies.cardRepository, cardMetrics)
			output, err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.input)

			scenario.expect(output, err)
		})
	}
}

func intPtrUC(v int) *int {
	return &v
}
