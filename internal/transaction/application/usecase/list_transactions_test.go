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
	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	transactionInterfaces "github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
	transactionMocks "github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces/mocks"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
)

type ListTransactionsUseCaseSuite struct {
	suite.Suite
	ctx  context.Context
	obs  *fake.Provider
	repo *transactionMocks.TransactionRepository
}

func TestListTransactionsUseCaseSuite(t *testing.T) {
	suite.Run(t, new(ListTransactionsUseCaseSuite))
}

func (s *ListTransactionsUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.repo = transactionMocks.NewTransactionRepository(s.T())
}

func makeTransactions(userID string, count int) []*entities.Transaction {
	userUUID, _ := vos.NewUUIDFromString(userID)
	catUUID, _ := vos.NewUUID()
	amount, _ := vos.NewMoneyFromFloat(50.0, vos.CurrencyBRL)
	pm, _ := transactionVos.NewPaymentMethod("pix")
	status, _ := transactionVos.NewTransactionStatus(transactionVos.TransactionStatusActive)
	result := make([]*entities.Transaction, count)
	for i := range result {
		uid, _ := vos.NewUUID()
		tx, _ := entities.NewTransaction(entities.TransactionParams{
			ID:              uid,
			UserID:          userUUID,
			CategoryID:      catUUID,
			Description:     "Item",
			Amount:          amount,
			PaymentMethod:   pm,
			TransactionDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			Status:          status,
			CreatedAt:       time.Now().UTC(),
		})
		result[i] = tx
	}
	return result
}

func (s *ListTransactionsUseCaseSuite) TestExecute() {
	userID := "550e8400-e29b-41d4-a716-446655440000"

	type args struct {
		userID string
		params *dtos.ListParams
	}
	type dependencies func()
	type expect func(output *dtos.TransactionListOutput, err error)

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       expect
	}{
		{
			name: "should return paginated list without filters",
			args: args{
				userID: userID,
				params: &dtos.ListParams{Limit: 20},
			},
			dependencies: func() {
				txs := makeTransactions(userID, 3)
				s.repo.EXPECT().ListPaginated(mock.Anything, mock.MatchedBy(func(p transactionInterfaces.ListParams) bool {
					return p.Limit == 20
				})).Return(txs, "", nil).Once()
			},
			expect: func(output *dtos.TransactionListOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Data, 3)
				s.Empty(output.NextCursor)
			},
		},
		{
			name: "should filter by payment_method",
			args: args{
				userID: userID,
				params: &dtos.ListParams{PaymentMethod: "pix", Limit: 10},
			},
			dependencies: func() {
				txs := makeTransactions(userID, 2)
				s.repo.EXPECT().ListPaginated(mock.Anything, mock.MatchedBy(func(p transactionInterfaces.ListParams) bool {
					return p.PaymentMethod == "pix" && p.Limit == 10
				})).Return(txs, "", nil).Once()
			},
			expect: func(output *dtos.TransactionListOutput, err error) {
				s.NoError(err)
				s.Len(output.Data, 2)
			},
		},
		{
			name: "should clamp limit to 100 when above max",
			args: args{
				userID: userID,
				params: &dtos.ListParams{Limit: 200},
			},
			dependencies: func() {
				txs := makeTransactions(userID, 1)
				s.repo.EXPECT().ListPaginated(mock.Anything, mock.MatchedBy(func(p transactionInterfaces.ListParams) bool {
					return p.Limit == 100
				})).Return(txs, "", nil).Once()
			},
			expect: func(output *dtos.TransactionListOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Len(output.Data, 1)
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.dependencies()
			uc := NewListTransactionsUseCase(s.obs, s.repo)
			output, err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.params)
			scenario.expect(output, err)
		})
	}
}
