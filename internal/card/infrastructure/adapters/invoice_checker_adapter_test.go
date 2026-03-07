package adapters_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/JailtonJunior94/devkit-go/pkg/observability/fake"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/card/infrastructure/adapters"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/entities"
	invoiceInterfaces "github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

type mockInvoiceRepository struct {
	mock.Mock
}

func (m *mockInvoiceRepository) FindByCard(ctx context.Context, cardID vos.UUID) ([]*entities.Invoice, error) {
	args := m.Called(ctx, cardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Invoice), args.Error(1)
}

func (m *mockInvoiceRepository) Insert(ctx context.Context, invoice *entities.Invoice) error {
	panic("not implemented")
}

func (m *mockInvoiceRepository) UpsertInvoice(ctx context.Context, invoice *entities.Invoice) (*entities.Invoice, error) {
	panic("not implemented")
}

func (m *mockInvoiceRepository) InsertItems(ctx context.Context, items []*entities.InvoiceItem) error {
	panic("not implemented")
}

func (m *mockInvoiceRepository) FindByID(ctx context.Context, id vos.UUID) (*entities.Invoice, error) {
	panic("not implemented")
}

func (m *mockInvoiceRepository) FindByUserAndCardAndMonth(ctx context.Context, userID vos.UUID, cardID vos.UUID, referenceMonth pkgVos.ReferenceMonth) (*entities.Invoice, error) {
	panic("not implemented")
}

func (m *mockInvoiceRepository) FindByUserAndMonth(ctx context.Context, userID vos.UUID, referenceMonth pkgVos.ReferenceMonth) ([]*entities.Invoice, error) {
	panic("not implemented")
}

func (m *mockInvoiceRepository) ListByCard(ctx context.Context, params invoiceInterfaces.ListInvoicesByCardParams) ([]*entities.Invoice, error) {
	panic("not implemented")
}

func (m *mockInvoiceRepository) ListByUserAndMonthPaginated(ctx context.Context, params invoiceInterfaces.ListInvoicesByMonthParams) ([]*entities.Invoice, error) {
	panic("not implemented")
}

func (m *mockInvoiceRepository) Update(ctx context.Context, invoice *entities.Invoice) error {
	panic("not implemented")
}

func (m *mockInvoiceRepository) UpdateItem(ctx context.Context, item *entities.InvoiceItem) error {
	panic("not implemented")
}

func (m *mockInvoiceRepository) DeleteItem(ctx context.Context, itemID vos.UUID) error {
	panic("not implemented")
}

func (m *mockInvoiceRepository) FindItemsByPurchaseOrigin(ctx context.Context, purchaseDate string, categoryID vos.UUID, description string) ([]*entities.InvoiceItem, error) {
	panic("not implemented")
}

func (m *mockInvoiceRepository) FindStatus(ctx context.Context, invoiceID vos.UUID) (string, error) {
	panic("not implemented")
}

func TestInvoiceCheckerAdapter_HasOpenInvoices(t *testing.T) {
	cardID, _ := vos.NewUUIDFromString("550e8400-e29b-41d4-a716-446655440000")

	t.Run("should return false when no invoices", func(t *testing.T) {
		obs := fake.NewProvider()
		mockRepo := &mockInvoiceRepository{}
		mockRepo.On("FindByCard", mock.Anything, cardID).Return([]*entities.Invoice{}, nil).Once()

		checker := adapters.NewInvoiceCheckerAdapter(mockRepo, obs)
		hasOpen, err := checker.HasOpenInvoices(context.Background(), cardID)

		require.NoError(t, err)
		require.False(t, hasOpen)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return false when all invoices are past due", func(t *testing.T) {
		obs := fake.NewProvider()
		mockRepo := &mockInvoiceRepository{}
		pastInvoice := &entities.Invoice{DueDate: time.Now().Add(-24 * time.Hour)}
		mockRepo.On("FindByCard", mock.Anything, cardID).Return([]*entities.Invoice{pastInvoice}, nil).Once()

		checker := adapters.NewInvoiceCheckerAdapter(mockRepo, obs)
		hasOpen, err := checker.HasOpenInvoices(context.Background(), cardID)

		require.NoError(t, err)
		require.False(t, hasOpen)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return true when invoice has future due date", func(t *testing.T) {
		obs := fake.NewProvider()
		mockRepo := &mockInvoiceRepository{}
		futureInvoice := &entities.Invoice{DueDate: time.Now().Add(24 * time.Hour)}
		mockRepo.On("FindByCard", mock.Anything, cardID).Return([]*entities.Invoice{futureInvoice}, nil).Once()

		checker := adapters.NewInvoiceCheckerAdapter(mockRepo, obs)
		hasOpen, err := checker.HasOpenInvoices(context.Background(), cardID)

		require.NoError(t, err)
		require.True(t, hasOpen)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return true when invoice due date is today", func(t *testing.T) {
		obs := fake.NewProvider()
		mockRepo := &mockInvoiceRepository{}
		todayInvoice := &entities.Invoice{DueDate: time.Now().Add(time.Second)}
		mockRepo.On("FindByCard", mock.Anything, cardID).Return([]*entities.Invoice{todayInvoice}, nil).Once()

		checker := adapters.NewInvoiceCheckerAdapter(mockRepo, obs)
		hasOpen, err := checker.HasOpenInvoices(context.Background(), cardID)

		require.NoError(t, err)
		require.True(t, hasOpen)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		obs := fake.NewProvider()
		mockRepo := &mockInvoiceRepository{}
		mockRepo.On("FindByCard", mock.Anything, cardID).Return(nil, errors.New("db error")).Once()

		checker := adapters.NewInvoiceCheckerAdapter(mockRepo, obs)
		hasOpen, err := checker.HasOpenInvoices(context.Background(), cardID)

		require.Error(t, err)
		require.False(t, hasOpen)
		mockRepo.AssertExpectations(t)
	})
}
