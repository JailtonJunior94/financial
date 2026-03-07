package dtos_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
)

func validPixInput() *dtos.TransactionInput {
	return &dtos.TransactionInput{
		Description:     "Supermercado",
		Amount:          100.00,
		PaymentMethod:   "pix",
		TransactionDate: time.Now().Add(-24 * time.Hour).Format("2006-01-02"),
		CategoryID:      "01965b87-b35a-7f18-a3b1-000000000001",
	}
}

func TestTransactionInput_Validate(t *testing.T) {
	t.Run("should pass with all valid fields (pix)", func(t *testing.T) {
		err := validPixInput().Validate()
		require.NoError(t, err)
	})

	t.Run("should return error for credit without card_id", func(t *testing.T) {
		input := validPixInput()
		input.PaymentMethod = "credit"
		err := input.Validate()
		require.Error(t, err)
	})

	t.Run("should return error for pix with card_id set", func(t *testing.T) {
		input := validPixInput()
		input.CardID = "01965b87-b35a-7f18-a3b1-000000000003"
		err := input.Validate()
		require.Error(t, err)
	})

	t.Run("should return error for pix with installments > 1", func(t *testing.T) {
		input := validPixInput()
		input.Installments = 2
		err := input.Validate()
		require.Error(t, err)
	})

	t.Run("should return error for credit with installments = 49", func(t *testing.T) {
		input := validPixInput()
		input.PaymentMethod = "credit"
		input.CardID = "01965b87-b35a-7f18-a3b1-000000000003"
		input.Installments = 49
		err := input.Validate()
		require.Error(t, err)
	})

	t.Run("should return error for transaction_date in future", func(t *testing.T) {
		input := validPixInput()
		input.TransactionDate = time.Now().Add(48 * time.Hour).Format("2006-01-02")
		err := input.Validate()
		require.Error(t, err)
	})

	t.Run("should return error for amount = 0", func(t *testing.T) {
		input := validPixInput()
		input.Amount = 0
		err := input.Validate()
		require.Error(t, err)
	})

	t.Run("should return error for negative amount", func(t *testing.T) {
		input := validPixInput()
		input.Amount = -10.00
		err := input.Validate()
		require.Error(t, err)
	})

	t.Run("should return error for empty description", func(t *testing.T) {
		input := validPixInput()
		input.Description = ""
		err := input.Validate()
		require.Error(t, err)
	})

	t.Run("should pass for credit + installments=1 + card_id", func(t *testing.T) {
		input := validPixInput()
		input.PaymentMethod = "credit"
		input.CardID = "01965b87-b35a-7f18-a3b1-000000000003"
		input.Installments = 1
		err := input.Validate()
		require.NoError(t, err)
	})
}
