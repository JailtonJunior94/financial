package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financial/internal/invoice/domain/entities"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	invoiceMocks "github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces/mocks"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

type ListInvoicesByCardPaginatedUseCaseSuite struct {
	suite.Suite
	ctx  context.Context
	obs  *fake.Provider
	repo *invoiceMocks.InvoiceRepository
}

func TestListInvoicesByCardPaginatedUseCaseSuite(t *testing.T) {
	suite.Run(t, new(ListInvoicesByCardPaginatedUseCaseSuite))
}

func (s *ListInvoicesByCardPaginatedUseCaseSuite) SetupTest() {
	s.obs = fake.NewProvider()
	s.ctx = context.Background()
	s.repo = invoiceMocks.NewInvoiceRepository(s.T())
}

func (s *ListInvoicesByCardPaginatedUseCaseSuite) TestExecute_WithStatusFilter_ShouldPassStatusToRepository() {
	userID, _ := vos.NewUUID()
	cardID, _ := vos.NewUUID()
	invoiceID, _ := vos.NewUUID()
	refMonth := pkgVos.NewReferenceMonthFromDate(time.Now())

	invoice := entities.NewInvoice(userID, cardID, refMonth, time.Now(), vos.CurrencyBRL)
	invoice.SetID(invoiceID)

	s.repo.EXPECT().
		ListByCard(mock.Anything, mock.MatchedBy(func(p interfaces.ListInvoicesByCardParams) bool {
			return p.Status == "open" && p.UserID == userID && p.CardID == cardID
		})).
		Return([]*entities.Invoice{invoice}, nil).
		Once()

	uc := NewListInvoicesByCardPaginatedUseCase(s.repo, s.obs)
	output, err := uc.Execute(s.ctx, ListInvoicesByCardPaginatedInput{
		UserID: userID.String(),
		CardID: cardID.String(),
		Status: "open",
		Limit:  20,
	})

	s.NoError(err)
	s.NotNil(output)
	s.Len(output.Invoices, 1)
}

func (s *ListInvoicesByCardPaginatedUseCaseSuite) TestExecute_WithoutStatusFilter_ShouldPassEmptyStatus() {
	userID, _ := vos.NewUUID()
	cardID, _ := vos.NewUUID()

	s.repo.EXPECT().
		ListByCard(mock.Anything, mock.MatchedBy(func(p interfaces.ListInvoicesByCardParams) bool {
			return p.Status == "" && p.UserID == userID && p.CardID == cardID
		})).
		Return([]*entities.Invoice{}, nil).
		Once()

	uc := NewListInvoicesByCardPaginatedUseCase(s.repo, s.obs)
	output, err := uc.Execute(s.ctx, ListInvoicesByCardPaginatedInput{
		UserID: userID.String(),
		CardID: cardID.String(),
		Limit:  20,
	})

	s.NoError(err)
	s.NotNil(output)
	s.Empty(output.Invoices)
}
