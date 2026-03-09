package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/jailtonjunior94/financial/internal/card/application/dtos"
	cardDomain "github.com/jailtonjunior94/financial/internal/card/domain"
	repositoryMock "github.com/jailtonjunior94/financial/internal/card/infrastructure/repositories/mocks"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

type UpdateCardUseCaseSuite struct {
	suite.Suite

	ctx  context.Context
	obs  observability.Observability
	repo *repositoryMock.CardRepository
}

func TestUpdateCardUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UpdateCardUseCaseSuite))
}

func (s *UpdateCardUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.repo = repositoryMock.NewCardRepository(s.T())
}

func (s *UpdateCardUseCaseSuite) TestExecute() {
	const validUserID = "550e8400-e29b-41d4-a716-446655440000"
	const validCardID = "660e8400-e29b-41d4-a716-446655440001"

	type args struct {
		userID string
		cardID string
		input  *dtos.CardUpdateInput
	}

	type dependencies struct {
		setupMocks func()
	}

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       func(output *dtos.CardOutput, err error)
	}{
		{
			name: "should update credit card successfully",
			args: args{
				userID: validUserID,
				cardID: validCardID,
				input: &dtos.CardUpdateInput{
					Name:           "Updated Nubank",
					Flag:           "visa",
					LastFourDigits: "9999",
					DueDay:         intPtrUC(20),
				},
			},
			dependencies: dependencies{
				setupMocks: func() {
					creditCard := buildCreditCard(s.T(), validUserID)
					s.repo.EXPECT().FindByIDOnly(mock.Anything, mock.AnythingOfType("vos.UUID")).Return(creditCard, nil).Once()
					s.repo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*entities.Card")).Return(nil).Once()
				},
			},
			expect: func(output *dtos.CardOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal("Updated Nubank", output.Name)
				s.Equal("credit", output.Type)
				s.NotNil(output.DueDay)
			},
		},
		{
			name: "should update debit card successfully",
			args: args{
				userID: validUserID,
				cardID: validCardID,
				input: &dtos.CardUpdateInput{
					Name:           "Updated Debit",
					Flag:           "mastercard",
					LastFourDigits: "1111",
				},
			},
			dependencies: dependencies{
				setupMocks: func() {
					debitCard := buildDebitCard(s.T(), validUserID)
					s.repo.EXPECT().FindByIDOnly(mock.Anything, mock.AnythingOfType("vos.UUID")).Return(debitCard, nil).Once()
					s.repo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*entities.Card")).Return(nil).Once()
				},
			},
			expect: func(output *dtos.CardOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal("debit", output.Type)
				s.Nil(output.DueDay)
			},
		},
		{
			name: "should return error when card not found",
			args: args{
				userID: validUserID,
				cardID: validCardID,
				input:  &dtos.CardUpdateInput{Name: "Name", Flag: "visa", LastFourDigits: "1234"},
			},
			dependencies: dependencies{
				setupMocks: func() {
					s.repo.EXPECT().FindByIDOnly(mock.Anything, mock.AnythingOfType("vos.UUID")).Return(nil, nil).Once()
				},
			},
			expect: func(output *dtos.CardOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.ErrorIs(err, cardDomain.ErrCardNotFound)
			},
		},
		{
			name: "should return error when repository find fails",
			args: args{
				userID: validUserID,
				cardID: validCardID,
				input:  &dtos.CardUpdateInput{Name: "Name", Flag: "visa", LastFourDigits: "1234"},
			},
			dependencies: dependencies{
				setupMocks: func() {
					s.repo.EXPECT().FindByIDOnly(mock.Anything, mock.AnythingOfType("vos.UUID")).Return(nil, errors.New("db error")).Once()
				},
			},
			expect: func(output *dtos.CardOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "db error")
			},
		},
		{
			name: "should return error when repository update fails",
			args: args{
				userID: validUserID,
				cardID: validCardID,
				input:  &dtos.CardUpdateInput{Name: "Name", Flag: "visa", LastFourDigits: "1234"},
			},
			dependencies: dependencies{
				setupMocks: func() {
					debitCard := buildDebitCard(s.T(), validUserID)
					s.repo.EXPECT().FindByIDOnly(mock.Anything, mock.AnythingOfType("vos.UUID")).Return(debitCard, nil).Once()
					s.repo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*entities.Card")).Return(errors.New("update failed")).Once()
				},
			},
			expect: func(output *dtos.CardOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "update failed")
			},
		},
		{
			name: "should return error when user id is invalid",
			args: args{
				userID: "invalid-uuid",
				cardID: validCardID,
				input:  &dtos.CardUpdateInput{Name: "Name", Flag: "visa", LastFourDigits: "1234"},
			},
			dependencies: dependencies{
				setupMocks: func() {},
			},
			expect: func(output *dtos.CardOutput, err error) {
				s.Error(err)
				s.Nil(output)
			},
		},
		{
			name: "should return forbidden when card belongs to another user",
			args: args{
				userID: validUserID,
				cardID: validCardID,
				input:  &dtos.CardUpdateInput{Name: "Name", Flag: "visa", LastFourDigits: "1234"},
			},
			dependencies: dependencies{
				setupMocks: func() {
					otherUserID := "770e8400-e29b-41d4-a716-446655440099"
					cardOfAnotherUser := buildCreditCard(s.T(), otherUserID)
					s.repo.EXPECT().FindByIDOnly(mock.Anything, mock.AnythingOfType("vos.UUID")).Return(cardOfAnotherUser, nil).Once()
				},
			},
			expect: func(output *dtos.CardOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.ErrorIs(err, customErrors.ErrForbidden)
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.dependencies.setupMocks()
			cardMetrics := metrics.NewTestCardMetrics()
			uc := NewUpdateCardUseCase(s.obs, s.repo, cardMetrics)
			output, err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.cardID, scenario.args.input)
			scenario.expect(output, err)
		})
	}
}
