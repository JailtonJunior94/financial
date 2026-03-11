package entities_test

import (
	"testing"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
)

func validTransactionParams(t *testing.T) entities.TransactionParams {
	t.Helper()
	id, _ := vos.NewUUID()
	userID, _ := vos.NewUUID()
	categoryID, _ := vos.NewUUID()
	invoiceID, _ := vos.NewUUID()
	amount, _ := vos.NewMoneyFromFloat(100.00, vos.CurrencyBRL)
	pm, _ := transactionVos.NewPaymentMethod(transactionVos.PaymentMethodCredit)
	status, _ := transactionVos.NewTransactionStatus(transactionVos.TransactionStatusActive)
	num := 1
	total := 1
	return entities.TransactionParams{
		ID:                id,
		UserID:            userID,
		CategoryID:        categoryID,
		InvoiceID:         &invoiceID,
		Description:       "Notebook",
		Amount:            amount,
		PaymentMethod:     pm,
		TransactionDate:   time.Now().Add(-time.Hour),
		InstallmentNumber: &num,
		InstallmentTotal:  &total,
		Status:            status,
		CreatedAt:         time.Now().UTC(),
	}
}

func TestNewTransaction(t *testing.T) {
	t.Run("should create transaction with credit payment and invoice_id", func(t *testing.T) {
		params := validTransactionParams(t)
		tx, err := entities.NewTransaction(params)
		require.NoError(t, err)
		require.NotNil(t, tx)
		require.Equal(t, "Notebook", tx.Description)
	})

	t.Run("should create transaction with pix payment without card_id", func(t *testing.T) {
		params := validTransactionParams(t)
		pm, _ := transactionVos.NewPaymentMethod(transactionVos.PaymentMethodPix)
		params.PaymentMethod = pm
		params.CardID = nil
		params.InvoiceID = nil
		tx, err := entities.NewTransaction(params)
		require.NoError(t, err)
		require.NotNil(t, tx)
		require.Nil(t, tx.CardID)
		require.Nil(t, tx.InvoiceID)
	})

	t.Run("should return error when description is empty", func(t *testing.T) {
		params := validTransactionParams(t)
		params.Description = ""
		_, err := entities.NewTransaction(params)
		require.Error(t, err)
	})
}

func TestTransaction_Cancel(t *testing.T) {
	t.Run("should set status to cancelled and update UpdatedAt", func(t *testing.T) {
		params := validTransactionParams(t)
		tx, err := entities.NewTransaction(params)
		require.NoError(t, err)
		err = tx.Cancel()
		require.NoError(t, err)
		require.True(t, tx.Status.IsCancelled())
		require.NotNil(t, tx.UpdatedAt)
	})
}

func TestTransaction_UpdateDetails(t *testing.T) {
	t.Run("should update fields successfully", func(t *testing.T) {
		params := validTransactionParams(t)
		tx, err := entities.NewTransaction(params)
		require.NoError(t, err)
		newAmount, _ := vos.NewMoneyFromFloat(200.00, vos.CurrencyBRL)
		newCategoryID, _ := vos.NewUUID()
		err = tx.UpdateDetails("Updated desc", newAmount, newCategoryID)
		require.NoError(t, err)
		require.Equal(t, "Updated desc", tx.Description)
		require.Equal(t, int64(20000), tx.Amount.Cents())
		require.NotNil(t, tx.UpdatedAt)
	})

	t.Run("should return error when amount is zero", func(t *testing.T) {
		params := validTransactionParams(t)
		tx, err := entities.NewTransaction(params)
		require.NoError(t, err)
		zeroAmount, _ := vos.NewMoney(0, vos.CurrencyBRL)
		categoryID, _ := vos.NewUUID()
		err = tx.UpdateDetails("desc", zeroAmount, categoryID)
		require.Error(t, err)
	})
}

func TestTransaction_IsEditable(t *testing.T) {
	t.Run("should return true for open invoice", func(t *testing.T) {
		params := validTransactionParams(t)
		tx, _ := entities.NewTransaction(params)
		require.True(t, tx.IsEditable("open"))
	})

	t.Run("should return false for closed invoice", func(t *testing.T) {
		params := validTransactionParams(t)
		tx, _ := entities.NewTransaction(params)
		require.False(t, tx.IsEditable("closed"))
	})

	t.Run("should return false for paid invoice", func(t *testing.T) {
		params := validTransactionParams(t)
		tx, _ := entities.NewTransaction(params)
		require.False(t, tx.IsEditable("paid"))
	})

	t.Run("should return true for empty invoice status (non-credit transaction)", func(t *testing.T) {
		params := validTransactionParams(t)
		tx, _ := entities.NewTransaction(params)
		require.True(t, tx.IsEditable(""))
	})
}
