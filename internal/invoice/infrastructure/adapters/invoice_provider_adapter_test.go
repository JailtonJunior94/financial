package adapters_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	invoiceEntities "github.com/jailtonjunior94/financial/internal/invoice/domain/entities"
	invoiceMocks "github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces/mocks"
	"github.com/jailtonjunior94/financial/internal/invoice/infrastructure/adapters"
	transactionInterfaces "github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

type InvoiceProviderAdapterSuite struct {
	suite.Suite
	ctx  context.Context
	obs  *fake.Provider
	repo *invoiceMocks.InvoiceRepository
}

func TestInvoiceProviderAdapterSuite(t *testing.T) {
	suite.Run(t, new(InvoiceProviderAdapterSuite))
}

func (s *InvoiceProviderAdapterSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.repo = invoiceMocks.NewInvoiceRepository(s.T())
}

func (s *InvoiceProviderAdapterSuite) TestFindOrCreate() {
	userID, _ := vos.NewUUID()
	cardID, _ := vos.NewUUID()
	invoiceID, _ := vos.NewUUID()
	refMonth, _ := pkgVos.NewReferenceMonth("2026-03")
	dueDate := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)

	type dependencies func()
	type expect func(info *transactionInterfaces.InvoiceInfo, err error)

	scenarios := []struct {
		name         string
		dependencies dependencies
		expect       expect
	}{
		{
			name: "should return invoice info on success",
			dependencies: func() {
				zeroMoney, _ := vos.NewMoney(0, vos.CurrencyBRL)
				invoice := &invoiceEntities.Invoice{}
				invoice.SetID(invoiceID)
				invoice.UserID = userID
				invoice.CardID = cardID
				invoice.ReferenceMonth = refMonth
				invoice.DueDate = dueDate
				invoice.TotalAmount = zeroMoney
				invoice.Status = "open"
				s.repo.EXPECT().UpsertInvoice(mock.Anything, mock.Anything).Return(invoice, nil).Once()
			},
			expect: func(info *transactionInterfaces.InvoiceInfo, err error) {
				s.NoError(err)
				s.NotNil(info)
				s.Equal("open", info.Status)
			},
		},
		{
			name: "should propagate error from repository",
			dependencies: func() {
				s.repo.EXPECT().UpsertInvoice(mock.Anything, mock.Anything).Return(nil, errors.New("db error")).Once()
			},
			expect: func(info *transactionInterfaces.InvoiceInfo, err error) {
				s.Error(err)
				s.Nil(info)
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.dependencies()
			adapter := adapters.NewInvoiceProviderAdapter(s.repo, s.obs)
			info, err := adapter.FindOrCreate(s.ctx, userID, cardID, refMonth, dueDate)
			scenario.expect(info, err)
		})
	}
}

func (s *InvoiceProviderAdapterSuite) TestGetStatus() {
	invoiceID, _ := vos.NewUUID()

	type dependencies func()
	type expect func(status string, err error)

	scenarios := []struct {
		name         string
		dependencies dependencies
		expect       expect
	}{
		{
			name: "should return open status",
			dependencies: func() {
				s.repo.EXPECT().FindStatus(mock.Anything, mock.Anything).Return("open", nil).Once()
			},
			expect: func(status string, err error) {
				s.NoError(err)
				s.Equal("open", status)
			},
		},
		{
			name: "should return empty string when invoice not found",
			dependencies: func() {
				s.repo.EXPECT().FindStatus(mock.Anything, mock.Anything).Return("", nil).Once()
			},
			expect: func(status string, err error) {
				s.NoError(err)
				s.Equal("", status)
			},
		},
		{
			name: "should propagate error from repository",
			dependencies: func() {
				s.repo.EXPECT().FindStatus(mock.Anything, mock.Anything).Return("", errors.New("db error")).Once()
			},
			expect: func(status string, err error) {
				s.Error(err)
				s.Equal("", status)
			},
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			scenario.dependencies()
			adapter := adapters.NewInvoiceProviderAdapter(s.repo, s.obs)
			status, err := adapter.GetStatus(s.ctx, invoiceID)
			scenario.expect(status, err)
		})
	}
}
