package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/jailtonjunior94/financial/internal/card/domain/entities"
	repositoryMock "github.com/jailtonjunior94/financial/internal/card/infrastructure/repositories/mocks"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

type FindCardPaginatedUseCaseSuite struct {
	suite.Suite

	ctx  context.Context
	obs  observability.Observability
	repo *repositoryMock.CardRepository
}

func TestFindCardPaginatedUseCaseSuite(t *testing.T) {
	suite.Run(t, new(FindCardPaginatedUseCaseSuite))
}

func (s *FindCardPaginatedUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.repo = repositoryMock.NewCardRepository(s.T())
}

func (s *FindCardPaginatedUseCaseSuite) TestExecute() {
	const validUserID = "550e8400-e29b-41d4-a716-446655440000"

	type args struct {
		input FindCardPaginatedInput
	}

	type dependencies struct {
		setupMocks func()
	}

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       func(output *FindCardPaginatedOutput, err error)
	}{
		{
			name: "should return empty list when no cards",
			args: args{input: FindCardPaginatedInput{UserID: validUserID, Limit: 10}},
			dependencies: dependencies{
				setupMocks: func() {
					s.repo.EXPECT().ListPaginated(mock.Anything, mock.Anything).Return([]*entities.Card{}, nil).Once()
				},
			},
			expect: func(output *FindCardPaginatedOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Empty(output.Cards)
				s.Nil(output.NextCursor)
			},
		},
		{
			name: "should return cards without next cursor when results fit in one page",
			args: args{input: FindCardPaginatedInput{UserID: validUserID, Limit: 10}},
			dependencies: dependencies{
				setupMocks: func() {
					creditCard := buildCreditCard(s.T(), validUserID)
					s.repo.EXPECT().ListPaginated(mock.Anything, mock.Anything).Return([]*entities.Card{creditCard}, nil).Once()
				},
			},
			expect: func(output *FindCardPaginatedOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Cards, 1)
				s.Nil(output.NextCursor)
			},
		},
		{
			name: "should return next cursor when results exceed page limit",
			args: args{input: FindCardPaginatedInput{UserID: validUserID, Limit: 1}},
			dependencies: dependencies{
				setupMocks: func() {
					card1 := buildCreditCard(s.T(), validUserID)
					card2 := buildDebitCard(s.T(), validUserID)
					s.repo.EXPECT().ListPaginated(mock.Anything, mock.Anything).Return([]*entities.Card{card1, card2}, nil).Once()
				},
			},
			expect: func(output *FindCardPaginatedOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Cards, 1)
				s.NotNil(output.NextCursor)
			},
		},
		{
			name: "should return error when user id is invalid",
			args: args{input: FindCardPaginatedInput{UserID: "invalid-uuid", Limit: 10}},
			dependencies: dependencies{
				setupMocks: func() {},
			},
			expect: func(output *FindCardPaginatedOutput, err error) {
				s.Error(err)
				s.Nil(output)
			},
		},
		{
			name: "should return error when repository fails",
			args: args{input: FindCardPaginatedInput{UserID: validUserID, Limit: 10}},
			dependencies: dependencies{
				setupMocks: func() {
					s.repo.EXPECT().ListPaginated(mock.Anything, mock.Anything).Return(nil, errors.New("db error")).Once()
				},
			},
			expect: func(output *FindCardPaginatedOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.Contains(err.Error(), "db error")
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.dependencies.setupMocks()
			cardMetrics := metrics.NewTestCardMetrics()
			uc := NewFindCardPaginatedUseCase(s.obs, s.repo, cardMetrics)
			output, err := uc.Execute(s.ctx, scenario.args.input)
			scenario.expect(output, err)
		})
	}
}
