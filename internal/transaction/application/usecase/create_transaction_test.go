package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
	transactionDomain "github.com/jailtonjunior94/financial/internal/transaction/domain"
	transactionMocks "github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces/mocks"
	invoiceInterfaces "github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	invoiceMocks "github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces/mocks"
	transactionInterfaces "github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
	outboxMocks "github.com/jailtonjunior94/financial/pkg/outbox/mocks"
)

type mockUnitOfWork struct{}

func (m *mockUnitOfWork) Do(ctx context.Context, fn func(ctx context.Context, tx database.DBTX) error) error {
	return fn(ctx, nil)
}

type CreateTransactionUseCaseSuite struct {
	suite.Suite
	ctx             context.Context
	obs             *fake.Provider
	repo            *transactionMocks.TransactionRepository
	invoiceProvider *transactionMocks.InvoiceProvider
	cardProvider    *invoiceMocks.CardProvider
	outboxService   *outboxMocks.Service
}

func TestCreateTransactionUseCaseSuite(t *testing.T) {
	suite.Run(t, new(CreateTransactionUseCaseSuite))
}

func (s *CreateTransactionUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.repo = transactionMocks.NewTransactionRepository(s.T())
	s.invoiceProvider = transactionMocks.NewInvoiceProvider(s.T())
	s.cardProvider = invoiceMocks.NewCardProvider(s.T())
	s.outboxService = outboxMocks.NewService(s.T())
}

func (s *CreateTransactionUseCaseSuite) TestExecute() {
	validCardID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440010")
	validInvoiceID, _ := vos.NewUUID()
	billingInfo := &invoiceInterfaces.CardBillingInfo{
		CardID:            validCardID,
		DueDay:            10,
		ClosingOffsetDays: 3,
	}
	invoiceInfo := &transactionInterfaces.InvoiceInfo{
		ID:     validInvoiceID,
		Status: "open",
	}

	type args struct {
		userID string
		input  *dtos.TransactionInput
	}
	type dependencies func()
	type expect func(outputs []*dtos.TransactionOutput, err error)

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       expect
	}{
		{
			name: "should create pix transaction successfully",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.TransactionInput{
					Description:     "Lunch",
					Amount:          50.00,
					PaymentMethod:   "pix",
					TransactionDate: "2026-03-01",
					CategoryID:      "550e8400-e29b-41d4-a716-446655440001",
				},
			},
			dependencies: func() {
				s.repo.EXPECT().SaveAll(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				s.outboxService.EXPECT().SaveDomainEvent(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expect: func(outputs []*dtos.TransactionOutput, err error) {
				s.NoError(err)
				s.NotNil(outputs)
				s.Len(outputs, 1)
				s.Equal("pix", outputs[0].PaymentMethod)
				s.Nil(outputs[0].InvoiceID)
			},
		},
		{
			name: "should create credit transaction with 1 installment",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.TransactionInput{
					Description:     "Purchase",
					Amount:          100.00,
					PaymentMethod:   "credit",
					TransactionDate: "2026-03-01",
					CategoryID:      "550e8400-e29b-41d4-a716-446655440001",
					CardID:          "550e8400-e29b-41d4-a716-446655440010",
					Installments:    1,
				},
			},
			dependencies: func() {
				s.cardProvider.EXPECT().GetCardBillingInfo(mock.Anything, mock.Anything, mock.Anything).Return(billingInfo, nil).Once()
				s.invoiceProvider.EXPECT().FindOrCreate(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(invoiceInfo, nil).Once()
				s.repo.EXPECT().SaveAll(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				s.outboxService.EXPECT().SaveDomainEvent(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			expect: func(outputs []*dtos.TransactionOutput, err error) {
				s.NoError(err)
				s.NotNil(outputs)
				s.Len(outputs, 1)
				s.Equal("credit", outputs[0].PaymentMethod)
				s.NotNil(outputs[0].InvoiceID)
			},
		},
		{
			name: "should create credit transaction with 3 installments",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.TransactionInput{
					Description:     "TV",
					Amount:          900.00,
					PaymentMethod:   "credit",
					TransactionDate: "2026-03-01",
					CategoryID:      "550e8400-e29b-41d4-a716-446655440001",
					CardID:          "550e8400-e29b-41d4-a716-446655440010",
					Installments:    3,
				},
			},
			dependencies: func() {
				s.cardProvider.EXPECT().GetCardBillingInfo(mock.Anything, mock.Anything, mock.Anything).Return(billingInfo, nil).Once()
				s.invoiceProvider.EXPECT().FindOrCreate(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(invoiceInfo, nil).Times(3)
				s.repo.EXPECT().SaveAll(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				s.outboxService.EXPECT().SaveDomainEvent(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(3)
			},
			expect: func(outputs []*dtos.TransactionOutput, err error) {
				s.NoError(err)
				s.Len(outputs, 3)
				s.Equal(outputs[0].InstallmentGroupID, outputs[1].InstallmentGroupID)
				s.Equal(outputs[1].InstallmentGroupID, outputs[2].InstallmentGroupID)
			},
		},
		{
			name: "should return error when credit payment without card_id",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.TransactionInput{
					Description:     "Purchase",
					Amount:          100.00,
					PaymentMethod:   "credit",
					TransactionDate: "2026-03-01",
					CategoryID:      "550e8400-e29b-41d4-a716-446655440001",
				},
			},
			dependencies: func() {},
			expect: func(outputs []*dtos.TransactionOutput, err error) {
				s.Error(err)
				s.Nil(outputs)
				s.ErrorIs(err, transactionDomain.ErrCardRequiredForCredit)
			},
		},
		{
			name: "should return error when pix payment with card_id",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.TransactionInput{
					Description:     "Transfer",
					Amount:          50.00,
					PaymentMethod:   "pix",
					TransactionDate: "2026-03-01",
					CategoryID:      "550e8400-e29b-41d4-a716-446655440001",
					CardID:          "550e8400-e29b-41d4-a716-446655440010",
				},
			},
			dependencies: func() {},
			expect: func(outputs []*dtos.TransactionOutput, err error) {
				s.Error(err)
				s.Nil(outputs)
				s.ErrorIs(err, transactionDomain.ErrCardNotAllowedForMethod)
			},
		},
		{
			name: "should return error when installments on non-credit payment",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.TransactionInput{
					Description:     "Payment",
					Amount:          100.00,
					PaymentMethod:   "pix",
					TransactionDate: "2026-03-01",
					CategoryID:      "550e8400-e29b-41d4-a716-446655440001",
					Installments:    3,
				},
			},
			dependencies: func() {},
			expect: func(outputs []*dtos.TransactionOutput, err error) {
				s.Error(err)
				s.Nil(outputs)
				s.ErrorIs(err, transactionDomain.ErrInstallmentsOnlyForCredit)
			},
		},
		{
			name: "should return error when transaction_date is in the future",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.TransactionInput{
					Description:     "Future",
					Amount:          50.00,
					PaymentMethod:   "pix",
					TransactionDate: time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02"),
					CategoryID:      "550e8400-e29b-41d4-a716-446655440001",
				},
			},
			dependencies: func() {},
			expect: func(outputs []*dtos.TransactionOutput, err error) {
				s.Error(err)
				s.Nil(outputs)
				s.ErrorIs(err, transactionDomain.ErrTransactionDateFuture)
			},
		},
		{
			name: "should return error when installments exceed 48",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.TransactionInput{
					Description:     "Appliance",
					Amount:          5000.00,
					PaymentMethod:   "credit",
					TransactionDate: "2026-03-01",
					CategoryID:      "550e8400-e29b-41d4-a716-446655440001",
					CardID:          "550e8400-e29b-41d4-a716-446655440010",
					Installments:    49,
				},
			},
			dependencies: func() {},
			expect: func(outputs []*dtos.TransactionOutput, err error) {
				s.Error(err)
				s.Nil(outputs)
				s.ErrorIs(err, transactionDomain.ErrInstallmentsTooMany)
			},
		},
		{
			name: "should propagate error from card provider",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.TransactionInput{
					Description:     "Purchase",
					Amount:          100.00,
					PaymentMethod:   "credit",
					TransactionDate: "2026-03-01",
					CategoryID:      "550e8400-e29b-41d4-a716-446655440001",
					CardID:          "550e8400-e29b-41d4-a716-446655440010",
					Installments:    1,
				},
			},
			dependencies: func() {
				s.cardProvider.EXPECT().GetCardBillingInfo(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("card not found")).Once()
			},
			expect: func(outputs []*dtos.TransactionOutput, err error) {
				s.Error(err)
				s.Nil(outputs)
				s.Contains(err.Error(), "card not found")
			},
		},
		{
			name: "should propagate error from invoice provider",
			args: args{
				userID: "550e8400-e29b-41d4-a716-446655440000",
				input: &dtos.TransactionInput{
					Description:     "Purchase",
					Amount:          100.00,
					PaymentMethod:   "credit",
					TransactionDate: "2026-03-01",
					CategoryID:      "550e8400-e29b-41d4-a716-446655440001",
					CardID:          "550e8400-e29b-41d4-a716-446655440010",
					Installments:    1,
				},
			},
			dependencies: func() {
				s.cardProvider.EXPECT().GetCardBillingInfo(mock.Anything, mock.Anything, mock.Anything).Return(billingInfo, nil).Once()
				s.invoiceProvider.EXPECT().FindOrCreate(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("db error")).Once()
			},
			expect: func(outputs []*dtos.TransactionOutput, err error) {
				s.Error(err)
				s.Nil(outputs)
				s.Contains(err.Error(), "db error")
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.dependencies()
			uc := NewCreateTransactionUseCase(
				s.obs,
				&mockUnitOfWork{},
				s.repo,
				s.invoiceProvider,
				s.cardProvider,
				s.outboxService,
			)
			outputs, err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.input)
			scenario.expect(outputs, err)
		})
	}
}
