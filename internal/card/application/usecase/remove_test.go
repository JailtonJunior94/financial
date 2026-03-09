package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	domain "github.com/jailtonjunior94/financial/internal/card/domain"
	"github.com/jailtonjunior94/financial/internal/card/domain/entities"
	cardVos "github.com/jailtonjunior94/financial/internal/card/domain/vos"
	repositoryMock "github.com/jailtonjunior94/financial/internal/card/infrastructure/repositories/mocks"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

type RemoveCardUseCaseSuite struct {
	suite.Suite

	ctx            context.Context
	obs            observability.Observability
	repo           *repositoryMock.CardRepository
	invoiceChecker *repositoryMock.InvoiceChecker
	cardMetrics    *metrics.CardMetrics
}

func TestRemoveCardUseCaseSuite(t *testing.T) {
	suite.Run(t, new(RemoveCardUseCaseSuite))
}

func (s *RemoveCardUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.repo = repositoryMock.NewCardRepository(s.T())
	s.invoiceChecker = repositoryMock.NewInvoiceChecker(s.T())
	s.cardMetrics = metrics.NewTestCardMetrics()
}

func (s *RemoveCardUseCaseSuite) TestExecute() {
	const validUserID = "550e8400-e29b-41d4-a716-446655440000"
	const validCardID = "660e8400-e29b-41d4-a716-446655440001"

	type args struct {
		userID string
		cardID string
	}

	type dependencies struct {
		setupMocks func()
	}

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       func(err error)
	}{
		{
			name: "should remove debit card without checking invoices",
			args: args{userID: validUserID, cardID: validCardID},
			dependencies: dependencies{
				setupMocks: func() {
					debitCard := buildDebitCard(s.T(), validUserID)
					s.repo.EXPECT().FindByIDOnly(mock.Anything, mock.AnythingOfType("vos.UUID")).Return(debitCard, nil).Once()
					s.repo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*entities.Card")).Return(nil).Once()
				},
			},
			expect: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "should remove credit card without open invoices",
			args: args{userID: validUserID, cardID: validCardID},
			dependencies: dependencies{
				setupMocks: func() {
					creditCard := buildCreditCard(s.T(), validUserID)
					s.repo.EXPECT().FindByIDOnly(mock.Anything, mock.AnythingOfType("vos.UUID")).Return(creditCard, nil).Once()
					s.invoiceChecker.EXPECT().HasOpenInvoices(mock.Anything, mock.AnythingOfType("vos.UUID")).Return(false, nil).Once()
					s.repo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*entities.Card")).Return(nil).Once()
				},
			},
			expect: func(err error) {
				s.NoError(err)
			},
		},
		{
			name: "should return error when credit card has open invoices",
			args: args{userID: validUserID, cardID: validCardID},
			dependencies: dependencies{
				setupMocks: func() {
					creditCard := buildCreditCard(s.T(), validUserID)
					s.repo.EXPECT().FindByIDOnly(mock.Anything, mock.AnythingOfType("vos.UUID")).Return(creditCard, nil).Once()
					s.invoiceChecker.EXPECT().HasOpenInvoices(mock.Anything, mock.AnythingOfType("vos.UUID")).Return(true, nil).Once()
				},
			},
			expect: func(err error) {
				s.Error(err)
				s.ErrorIs(err, domain.ErrCardHasOpenInvoices)
			},
		},
		{
			name: "should return error when card not found",
			args: args{userID: validUserID, cardID: validCardID},
			dependencies: dependencies{
				setupMocks: func() {
					s.repo.EXPECT().FindByIDOnly(mock.Anything, mock.AnythingOfType("vos.UUID")).Return(nil, nil).Once()
				},
			},
			expect: func(err error) {
				s.Error(err)
				s.ErrorIs(err, domain.ErrCardNotFound)
			},
		},
		{
			name: "should return error when invoice checker fails",
			args: args{userID: validUserID, cardID: validCardID},
			dependencies: dependencies{
				setupMocks: func() {
					creditCard := buildCreditCard(s.T(), validUserID)
					s.repo.EXPECT().FindByIDOnly(mock.Anything, mock.AnythingOfType("vos.UUID")).Return(creditCard, nil).Once()
					s.invoiceChecker.EXPECT().HasOpenInvoices(mock.Anything, mock.AnythingOfType("vos.UUID")).Return(false, errors.New("database error")).Once()
				},
			},
			expect: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "database error")
			},
		},
		{
			name: "should return error when update fails",
			args: args{userID: validUserID, cardID: validCardID},
			dependencies: dependencies{
				setupMocks: func() {
					debitCard := buildDebitCard(s.T(), validUserID)
					s.repo.EXPECT().FindByIDOnly(mock.Anything, mock.AnythingOfType("vos.UUID")).Return(debitCard, nil).Once()
					s.repo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*entities.Card")).Return(errors.New("update failed")).Once()
				},
			},
			expect: func(err error) {
				s.Error(err)
				s.Contains(err.Error(), "update failed")
			},
		},
		{
			name: "should return forbidden when card belongs to another user",
			args: args{userID: validUserID, cardID: validCardID},
			dependencies: dependencies{
				setupMocks: func() {
					otherUserID := "770e8400-e29b-41d4-a716-446655440099"
					cardOfAnotherUser := buildDebitCard(s.T(), otherUserID)
					s.repo.EXPECT().FindByIDOnly(mock.Anything, mock.AnythingOfType("vos.UUID")).Return(cardOfAnotherUser, nil).Once()
				},
			},
			expect: func(err error) {
				s.Error(err)
				s.ErrorIs(err, customErrors.ErrForbidden)
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.dependencies.setupMocks()
			uc := NewRemoveCardUseCase(s.obs, s.repo, s.invoiceChecker, s.cardMetrics)
			err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.cardID)
			scenario.expect(err)
		})
	}
}

func buildCreditCard(t *testing.T, userIDStr string) *entities.Card {
	t.Helper()
	userID, _ := vos.NewUUIDFromString(userIDStr)
	name, _ := cardVos.NewCardName("Test Credit Card")
	cardType, _ := cardVos.NewCardType("credit")
	flag, _ := cardVos.NewCardFlag("visa")
	digits, _ := cardVos.NewLastFourDigits("1234")
	dueDay, _ := cardVos.NewDueDay(15)
	offset, _ := cardVos.NewClosingOffsetDays(7)
	card, _ := entities.NewCard(userID, name, cardType, flag, digits, dueDay, offset)
	cardID, _ := vos.NewUUID()
	card.ID = cardID
	return card
}

func buildDebitCard(t *testing.T, userIDStr string) *entities.Card {
	t.Helper()
	userID, _ := vos.NewUUIDFromString(userIDStr)
	name, _ := cardVos.NewCardName("Test Debit Card")
	cardType, _ := cardVos.NewCardType("debit")
	flag, _ := cardVos.NewCardFlag("mastercard")
	digits, _ := cardVos.NewLastFourDigits("5678")
	card, _ := entities.NewCard(userID, name, cardType, flag, digits, cardVos.DueDay{}, cardVos.ClosingOffsetDays{})
	cardID, _ := vos.NewUUID()
	card.ID = cardID
	return card
}
