package vos_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financial/internal/transaction/domain"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
)

func TestNewTransactionStatus(t *testing.T) {
	t.Run("should create active transaction status", func(t *testing.T) {
		status, err := vos.NewTransactionStatus("active")
		require.NoError(t, err)
		require.True(t, status.IsActive())
		require.False(t, status.IsCancelled())
		require.Equal(t, "active", status.String())
	})

	t.Run("should create cancelled transaction status", func(t *testing.T) {
		status, err := vos.NewTransactionStatus("cancelled")
		require.NoError(t, err)
		require.False(t, status.IsActive())
		require.True(t, status.IsCancelled())
		require.Equal(t, "cancelled", status.String())
	})

	t.Run("should return error for unknown status pending", func(t *testing.T) {
		_, err := vos.NewTransactionStatus("pending")
		require.ErrorIs(t, err, domain.ErrInvalidTransactionStatus)
	})

	t.Run("should return error for empty string", func(t *testing.T) {
		_, err := vos.NewTransactionStatus("")
		require.ErrorIs(t, err, domain.ErrInvalidTransactionStatus)
	})

	t.Run("should return error for uppercase Active", func(t *testing.T) {
		_, err := vos.NewTransactionStatus("Active")
		require.ErrorIs(t, err, domain.ErrInvalidTransactionStatus)
	})
}
