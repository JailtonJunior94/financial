package usecase

import (
	"context"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
	transactionDomain "github.com/jailtonjunior94/financial/internal/transaction/domain"
	transactionMocks "github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces/mocks"
)

type GetTransactionUseCaseSuite struct {
	suite.Suite
	ctx  context.Context
	obs  *fake.Provider
	repo *transactionMocks.TransactionRepository
}

func TestGetTransactionUseCaseSuite(t *testing.T) {
	suite.Run(t, new(GetTransactionUseCaseSuite))
}

func (s *GetTransactionUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.repo = transactionMocks.NewTransactionRepository(s.T())
}

func (s *GetTransactionUseCaseSuite) TestExecute() {
	userID := "550e8400-e29b-41d4-a716-446655440000"
	categoryID := "550e8400-e29b-41d4-a716-446655440001"

	type args struct {
		userID        string
		transactionID string
	}
	type dependencies func(txIDStr string)
	type expect func(output *dtos.TransactionOutput, err error)

	scenarios := []struct {
		name         string
		args         args
		dependencies dependencies
		expect       expect
	}{
		{
			name: "should return transaction when found and owned by user",
			args: args{userID: userID, transactionID: "660e8400-e29b-41d4-a716-446655440000"},
			dependencies: func(txIDStr string) {
				txID, _ := vos.NewUUIDFromString(txIDStr)
				tx := buildTransaction(userID, categoryID, nil)
				s.repo.EXPECT().FindByID(mock.Anything, txID).Return(tx, nil).Once()
			},
			expect: func(output *dtos.TransactionOutput, err error) {
				s.NoError(err)
				s.NotNil(output)
				s.Equal(userID, output.UserID)
			},
		},
		{
			name: "should return error when transaction belongs to another user",
			args: args{userID: "different-0000-0000-0000-000000000001", transactionID: "660e8400-e29b-41d4-a716-446655440000"},
			dependencies: func(txIDStr string) {
				txID, _ := vos.NewUUIDFromString(txIDStr)
				tx := buildTransaction(userID, categoryID, nil)
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
			args: args{userID: userID, transactionID: "660e8400-e29b-41d4-a716-446655440000"},
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
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.dependencies(scenario.args.transactionID)
			uc := NewGetTransactionUseCase(s.obs, s.repo)
			output, err := uc.Execute(s.ctx, scenario.args.userID, scenario.args.transactionID)
			scenario.expect(output, err)
		})
	}
}
