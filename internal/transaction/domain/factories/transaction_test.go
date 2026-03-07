package factories_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financial/internal/transaction/domain/factories"
)

const (
	testUserID     = "01965b87-b35a-7f18-a3b1-000000000001"
	testCategoryID = "01965b87-b35a-7f18-a3b1-000000000002"
	testCardID     = "01965b87-b35a-7f18-a3b1-000000000003"
	testInvoiceID  = "01965b87-b35a-7f18-a3b1-000000000004"
)

func baseCreateParams() factories.CreateParams {
	return factories.CreateParams{
		UserID:          testUserID,
		CategoryID:      testCategoryID,
		Description:     "Notebook",
		Amount:          100.00,
		PaymentMethod:   "pix",
		TransactionDate: time.Now().Add(-time.Hour),
		Installments:    1,
	}
}

func TestTransactionFactory_Create(t *testing.T) {
	factory := factories.NewTransactionFactory()

	t.Run("should create pix transaction without card_id", func(t *testing.T) {
		params := baseCreateParams()
		tx, err := factory.Create(params)
		require.NoError(t, err)
		require.NotNil(t, tx)
		require.Nil(t, tx.InvoiceID)
		require.Nil(t, tx.CardID)
	})

	t.Run("should create credit transaction with invoice_id", func(t *testing.T) {
		params := baseCreateParams()
		params.PaymentMethod = "credit"
		params.CardID = testCardID
		params.InvoiceID = testInvoiceID
		tx, err := factory.Create(params)
		require.NoError(t, err)
		require.NotNil(t, tx)
		require.NotNil(t, tx.InvoiceID)
		require.NotNil(t, tx.CardID)
	})
}

func TestTransactionFactory_CreateInstallments(t *testing.T) {
	factory := factories.NewTransactionFactory()

	t.Run("should split R$100 into 3 installments with correct rounding", func(t *testing.T) {
		params := factories.InstallmentParams{
			CreateParams: factories.CreateParams{
				UserID:          testUserID,
				CategoryID:      testCategoryID,
				CardID:          testCardID,
				Description:     "Notebook",
				Amount:          100.00,
				PaymentMethod:   "credit",
				TransactionDate: time.Now().Add(-time.Hour),
				Installments:    3,
			},
			InvoiceIDs: []string{testInvoiceID, testCardID, testUserID},
		}
		txs, err := factory.CreateInstallments(params)
		require.NoError(t, err)
		require.Len(t, txs, 3)
		require.Equal(t, int64(3333), txs[0].Amount.Cents())
		require.Equal(t, int64(3333), txs[1].Amount.Cents())
		require.Equal(t, int64(3334), txs[2].Amount.Cents())
	})

	t.Run("should split R$100.01 into 2 installments correctly", func(t *testing.T) {
		params := factories.InstallmentParams{
			CreateParams: factories.CreateParams{
				UserID:          testUserID,
				CategoryID:      testCategoryID,
				CardID:          testCardID,
				Description:     "Tablet",
				Amount:          100.01,
				PaymentMethod:   "credit",
				TransactionDate: time.Now().Add(-time.Hour),
				Installments:    2,
			},
			InvoiceIDs: []string{testInvoiceID, testCardID},
		}
		txs, err := factory.CreateInstallments(params)
		require.NoError(t, err)
		require.Len(t, txs, 2)
		require.Equal(t, int64(5000), txs[0].Amount.Cents())
		require.Equal(t, int64(5001), txs[1].Amount.Cents())
	})

	t.Run("should split R$1000 into 10 equal installments", func(t *testing.T) {
		invoiceIDs := make([]string, 10)
		for i := range invoiceIDs {
			invoiceIDs[i] = testInvoiceID
		}
		params := factories.InstallmentParams{
			CreateParams: factories.CreateParams{
				UserID:          testUserID,
				CategoryID:      testCategoryID,
				CardID:          testCardID,
				Description:     "TV",
				Amount:          1000.00,
				PaymentMethod:   "credit",
				TransactionDate: time.Now().Add(-time.Hour),
				Installments:    10,
			},
			InvoiceIDs: invoiceIDs,
		}
		txs, err := factory.CreateInstallments(params)
		require.NoError(t, err)
		require.Len(t, txs, 10)
		for _, tx := range txs {
			require.Equal(t, int64(10000), tx.Amount.Cents())
		}
	})

	t.Run("all installments share the same group ID", func(t *testing.T) {
		params := factories.InstallmentParams{
			CreateParams: factories.CreateParams{
				UserID:          testUserID,
				CategoryID:      testCategoryID,
				CardID:          testCardID,
				Description:     "Phone",
				Amount:          300.00,
				PaymentMethod:   "credit",
				TransactionDate: time.Now().Add(-time.Hour),
				Installments:    3,
			},
			InvoiceIDs: []string{testInvoiceID, testCardID, testUserID},
		}
		txs, err := factory.CreateInstallments(params)
		require.NoError(t, err)
		require.Len(t, txs, 3)
		groupID := txs[0].InstallmentGroupID
		require.NotNil(t, groupID)
		for _, tx := range txs {
			require.Equal(t, groupID.String(), tx.InstallmentGroupID.String())
		}
	})

	t.Run("installment numbers should be 1 to N in order", func(t *testing.T) {
		params := factories.InstallmentParams{
			CreateParams: factories.CreateParams{
				UserID:          testUserID,
				CategoryID:      testCategoryID,
				CardID:          testCardID,
				Description:     "Laptop",
				Amount:          600.00,
				PaymentMethod:   "credit",
				TransactionDate: time.Now().Add(-time.Hour),
				Installments:    3,
			},
			InvoiceIDs: []string{testInvoiceID, testCardID, testUserID},
		}
		txs, err := factory.CreateInstallments(params)
		require.NoError(t, err)
		require.Len(t, txs, 3)
		for i, tx := range txs {
			require.Equal(t, i+1, *tx.InstallmentNumber)
		}
	})

	t.Run("sum of all installments equals total amount", func(t *testing.T) {
		params := factories.InstallmentParams{
			CreateParams: factories.CreateParams{
				UserID:          testUserID,
				CategoryID:      testCategoryID,
				CardID:          testCardID,
				Description:     "Monitor",
				Amount:          100.00,
				PaymentMethod:   "credit",
				TransactionDate: time.Now().Add(-time.Hour),
				Installments:    3,
			},
			InvoiceIDs: []string{testInvoiceID, testCardID, testUserID},
		}
		txs, err := factory.CreateInstallments(params)
		require.NoError(t, err)
		var totalCents int64
		for _, tx := range txs {
			totalCents += tx.Amount.Cents()
		}
		require.Equal(t, int64(10000), totalCents)
	})
}
