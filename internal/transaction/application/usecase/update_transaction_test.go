package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
	transactionDomain "github.com/jailtonjunior94/financial/internal/transaction/domain"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	transactionMocks "github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces/mocks"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
)

type UpdateTransactionUseCaseSuite struct {
	suite.Suite
	ctx             context.Context
	obs             *fake.Provider
	repo            *transactionMocks.TransactionRepository
	invoiceProvider *transactionMocks.InvoiceProvider
}

func TestUpdateTransactionUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UpdateTransactionUseCaseSuite))
}

func (s *UpdateTransactionUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.repo = transactionMocks.NewTransactionRepository(s.T())
	s.invoiceProvider = transactionMocks.NewInvoiceProvider(s.T())
}

func buildTransaction(userID, categoryID string, invoiceID *vos.UUID) *entities.Transaction {
	uid, _ := vos.NewUUID()
	userUUID, _ := vos.NewUUIDFromString(userID)
	catUUID, _ := vos.NewUUIDFromString(categoryID)
	amount, _ := vos.NewMoneyFromFloat(100.0, vos.CurrencyBRL)
	pm, _ := transactionVos.NewPaymentMethod("credit")
	if invoiceID == nil {
		pm, _ = transactionVos.NewPaymentMethod("pix")
	}
	status, _ := transactionVos.NewTransactionStatus(transactionVos.TransactionStatusActive)
	tx, _ := entities.NewTransaction(entities.TransactionParams{
		ID:              uid,
		UserID:          userUUID,
		CategoryID:      catUUID,
		Description:     "Original",
		Amount:          amount,
		PaymentMethod:   pm,
		TransactionDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		Status:          status,
		CreatedAt:       time.Now().UTC(),
		InvoiceID:       invoiceID,
	})
	return tx
}

func (s *UpdateTransactionUseCaseSuite) TestExecute() {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	categoryID := "550e8400-e29b-41d4-a716-446655440001"
	invoiceID, _ := vos.NewUUID()

	type args struct {
		userID        string
		transactionID string
		input         *dtos.TransactionUpdateInput
	}
	type dependencies func(txID string)
	type expect func(output *dtos.TransactionOutput, err error)

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       expect
	}{
		{
			name: "should update transaction with open invoice",
			args: args{
				userID:        userID,
				transactionID: "660e8400-e29b-41d4-a716-446655440000",
				input: &dtos.TransactionUpdateInput{
					Description: "Updated",
					Amount:      200.00,
					CategoryID:  categoryID,
				},
			},
			dependencies: func(txIDStr string) {
				txID, _ := vos.NewUUIDFromString(txIDStr)
				tx := buildTransaction(userID, categoryID, &invoiceID)
				s.repo.EXPECT().FindByID(mock.Anything, txID).Return(tx, nil).Once()
				s.invoiceProvider.EXPECT().GetStatus(mock.Anything, invoiceID).Return("open", nil).Once()
				s.repo.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expect: func(output *dtos.TransactionOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal("Updated", output.Description)
				s.Equal(200.00, output.Amount)
			},
		},
		{
			name: "should return error when invoice is closed",
			args: args{
				userID:        userID,
				transactionID: "660e8400-e29b-41d4-a716-446655440000",
				input: &dtos.TransactionUpdateInput{
					Description: "Updated",
					Amount:      200.00,
					CategoryID:  categoryID,
				},
			},
			dependencies: func(txIDStr string) {
				txID, _ := vos.NewUUIDFromString(txIDStr)
				tx := buildTransaction(userID, categoryID, &invoiceID)
				s.repo.EXPECT().FindByID(mock.Anything, txID).Return(tx, nil).Once()
				s.invoiceProvider.EXPECT().GetStatus(mock.Anything, invoiceID).Return("closed", nil).Once()
			},
			expect: func(output *dtos.TransactionOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.ErrorIs(err, transactionDomain.ErrInvoiceClosed)
			},
		},
		{
			name: "should return error when transaction belongs to another user",
			args: args{
				userID:        "different-user-00-0000-0000-000000000001",
				transactionID: "660e8400-e29b-41d4-a716-446655440000",
				input: &dtos.TransactionUpdateInput{
					Description: "Hack",
					Amount:      200.00,
					CategoryID:  categoryID,
				},
			},
			dependencies: func(txIDStr string) {
				txID, _ := vos.NewUUIDFromString(txIDStr)
				tx := buildTransaction(userID, categoryID, &invoiceID)
				s.repo.EXPECT().FindByID(mock.Anything, txID).Return(tx, nil).Once()
			},
			expect: func(output *dtos.TransactionOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.ErrorIs(err, transactionDomain.ErrTransactionNotOwned)
			},
		},
		{
			name: "should return error when transaction not found",
			args: args{
				userID:        userID,
				transactionID: "660e8400-e29b-41d4-a716-446655440000",
				input: &dtos.TransactionUpdateInput{
					Description: "Updated",
					Amount:      200.00,
					CategoryID:  categoryID,
				},
			},
			dependencies: func(txIDStr string) {
				txID, _ := vos.NewUUIDFromString(txIDStr)
				s.repo.EXPECT().FindByID(mock.Anything, txID).Return(nil, nil).Once()
			},
			expect: func(output *dtos.TransactionOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.ErrorIs(err, transactionDomain.ErrTransactionNotFound)
			},
		},
		{
			name: "should update pix transaction without checking invoice status",
			args: args{
				userID:        userID,
				transactionID: "660e8400-e29b-41d4-a716-446655440000",
				input: &dtos.TransactionUpdateInput{
					Description: "Updated PIX",
					Amount:      75.00,
					CategoryID:  categoryID,
				},
			},
			dependencies: func(txIDStr string) {
				txID, _ := vos.NewUUIDFromString(txIDStr)
				tx := buildTransaction(userID, categoryID, nil)
				s.repo.EXPECT().FindByID(mock.Anything, txID).Return(tx, nil).Once()
				s.repo.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expect: func(output *dtos.TransactionOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal("Updated PIX", output.Description)
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.dependencies(scenario.args.transactionID)
			uc := NewUpdateTransactionUseCase(
				s.obs,
				&mockUnitOfWork{},
				s.repo,
				s.invoiceProvider,
			)
			output, err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.transactionID, scenario.args.input)
			scenario.expect(output, err)
		})
	}
}

func (s *UpdateTransactionUseCaseSuite) TestExecuteRepoError() {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	categoryID := "550e8400-e29b-41d4-a716-446655440001"
	txIDStr := "660e8400-e29b-41d4-a716-446655440000"
	txID, _ := vos.NewUUIDFromString(txIDStr)
	s.repo.EXPECT().FindByID(mock.Anything, txID).Return(nil, errors.New("db error")).Once()

	uc := NewUpdateTransactionUseCase(s.obs, &mockUnitOfWork{}, s.repo, s.invoiceProvider)
	output, err := uc.Execute(s.ctx, userID, txIDStr, &dtos.TransactionUpdateInput{
		Description: "Updated",
		Amount:      200.00,
		CategoryID:  categoryID,
	})
	s.Error(err)
	s.Nil(output)
}
