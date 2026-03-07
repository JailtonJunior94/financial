package usecase

import (
	"context"
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

type ReverseTransactionUseCaseSuite struct {
	suite.Suite
	ctx             context.Context
	obs             *fake.Provider
	repo            *transactionMocks.TransactionRepository
	invoiceProvider *transactionMocks.InvoiceProvider
}

func TestReverseTransactionUseCaseSuite(t *testing.T) {
	suite.Run(t, new(ReverseTransactionUseCaseSuite))
}

func (s *ReverseTransactionUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.repo = transactionMocks.NewTransactionRepository(s.T())
	s.invoiceProvider = transactionMocks.NewInvoiceProvider(s.T())
}

func makeTransaction(userID string, invoiceID *vos.UUID, groupID *vos.UUID) *entities.Transaction {
	uid, _ := vos.NewUUID()
	userUUID, _ := vos.NewUUIDFromString(userID)
	catUUID, _ := vos.NewUUID()
	amount, _ := vos.NewMoneyFromFloat(100.0, vos.CurrencyBRL)
	pm, _ := transactionVos.NewPaymentMethod("pix")
	if invoiceID != nil {
		pm, _ = transactionVos.NewPaymentMethod("credit")
	}
	status, _ := transactionVos.NewTransactionStatus(transactionVos.TransactionStatusActive)
	tx, _ := entities.NewTransaction(entities.TransactionParams{
		ID:                 uid,
		UserID:             userUUID,
		CategoryID:         catUUID,
		Description:        "Test",
		Amount:             amount,
		PaymentMethod:      pm,
		TransactionDate:    time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		Status:             status,
		CreatedAt:          time.Now().UTC(),
		InvoiceID:          invoiceID,
		InstallmentGroupID: groupID,
	})
	return tx
}

func (s *ReverseTransactionUseCaseSuite) TestExecute() {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	invoiceID, _ := vos.NewUUID()

	type args struct {
		userID        string
		transactionID string
	}
	type dependencies func(txIDStr string)
	type expect func(output *dtos.ReverseOutput, err error)

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       expect
	}{
		{
			name: "should cancel simple transaction with open invoice",
			args: args{userID: userID, transactionID: "660e8400-e29b-41d4-a716-446655440000"},
			dependencies: func(txIDStr string) {
				txID, _ := vos.NewUUIDFromString(txIDStr)
				tx := makeTransaction(userID, &invoiceID, nil)
				tx.ID = txID
				s.repo.EXPECT().FindByID(mock.Anything, txID).Return(tx, nil).Once()
				s.invoiceProvider.EXPECT().GetStatus(mock.Anything, invoiceID).Return("open", nil).Once()
				s.repo.EXPECT().UpdateAll(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expect: func(output *dtos.ReverseOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Cancelled, 1)
				s.Len(output.Kept, 0)
			},
		},
		{
			name: "should cancel open installments and keep closed ones",
			args: args{userID: userID, transactionID: "660e8400-e29b-41d4-a716-446655440001"},
			dependencies: func(txIDStr string) {
				txID, _ := vos.NewUUIDFromString(txIDStr)
				groupID, _ := vos.NewUUID()
				firstTx := makeTransaction(userID, &invoiceID, &groupID)
				firstTx.ID = txID
				s.repo.EXPECT().FindByID(mock.Anything, txID).Return(firstTx, nil).Once()

				allInstallments := make([]*entities.Transaction, 10)
				for i := 0; i < 10; i++ {
					invID, _ := vos.NewUUID()
					allInstallments[i] = makeTransaction(userID, &invID, &groupID)
				}
				s.repo.EXPECT().FindByInstallmentGroup(mock.Anything, groupID).Return(allInstallments, nil).Once()

				for i := 0; i < 10; i++ {
					inv := *allInstallments[i].InvoiceID
					status := "open"
					if i < 3 {
						status = "closed"
					}
					s.invoiceProvider.EXPECT().GetStatus(mock.Anything, inv).Return(status, nil).Once()
				}
				s.repo.EXPECT().UpdateAll(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expect: func(output *dtos.ReverseOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Cancelled, 7)
				s.Len(output.Kept, 3)
			},
		},
		{
			name: "should return error when all invoices are closed",
			args: args{userID: userID, transactionID: "660e8400-e29b-41d4-a716-446655440002"},
			dependencies: func(txIDStr string) {
				txID, _ := vos.NewUUIDFromString(txIDStr)
				groupID, _ := vos.NewUUID()
				firstTx := makeTransaction(userID, &invoiceID, &groupID)
				firstTx.ID = txID
				s.repo.EXPECT().FindByID(mock.Anything, txID).Return(firstTx, nil).Once()

				allInstallments := make([]*entities.Transaction, 3)
				for i := 0; i < 3; i++ {
					invID, _ := vos.NewUUID()
					allInstallments[i] = makeTransaction(userID, &invID, &groupID)
				}
				s.repo.EXPECT().FindByInstallmentGroup(mock.Anything, groupID).Return(allInstallments, nil).Once()

				for i := 0; i < 3; i++ {
					inv := *allInstallments[i].InvoiceID
					s.invoiceProvider.EXPECT().GetStatus(mock.Anything, inv).Return("closed", nil).Once()
				}
			},
			expect: func(output *dtos.ReverseOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.ErrorIs(err, transactionDomain.ErrNothingToReverse)
			},
		},
		{
			name: "should return error when transaction belongs to another user",
			args: args{userID: "different-0000-0000-0000-000000000001", transactionID: "660e8400-e29b-41d4-a716-446655440003"},
			dependencies: func(txIDStr string) {
				txID, _ := vos.NewUUIDFromString(txIDStr)
				tx := makeTransaction(userID, &invoiceID, nil)
				tx.ID = txID
				s.repo.EXPECT().FindByID(mock.Anything, txID).Return(tx, nil).Once()
			},
			expect: func(output *dtos.ReverseOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.ErrorIs(err, transactionDomain.ErrTransactionNotOwned)
			},
		},
		{
			name: "should return error when transaction not found",
			args: args{userID: userID, transactionID: "660e8400-e29b-41d4-a716-446655440004"},
			dependencies: func(txIDStr string) {
				txID, _ := vos.NewUUIDFromString(txIDStr)
				s.repo.EXPECT().FindByID(mock.Anything, txID).Return(nil, nil).Once()
			},
			expect: func(output *dtos.ReverseOutput, err error) {
				s.Error(err)
				s.Nil(output)
				s.ErrorIs(err, transactionDomain.ErrTransactionNotFound)
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.dependencies(scenario.args.transactionID)
			uc := NewReverseTransactionUseCase(
				s.obs,
				&mockUnitOfWork{},
				s.repo,
				s.invoiceProvider,
			)
			output, err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.transactionID)
			scenario.expect(output, err)
		})
	}
}
