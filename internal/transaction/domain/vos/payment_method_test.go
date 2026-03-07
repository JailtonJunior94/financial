package vos_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financial/internal/transaction/domain"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
)

func TestNewPaymentMethod(t *testing.T) {
	t.Run("should create pix payment method", func(t *testing.T) {
		pm, err := vos.NewPaymentMethod("pix")
		require.NoError(t, err)
		require.False(t, pm.IsCredit())
		require.False(t, pm.RequiresCard())
		require.Equal(t, "pix", pm.String())
	})

	t.Run("should create boleto payment method", func(t *testing.T) {
		pm, err := vos.NewPaymentMethod("boleto")
		require.NoError(t, err)
		require.False(t, pm.IsCredit())
		require.False(t, pm.RequiresCard())
		require.Equal(t, "boleto", pm.String())
	})

	t.Run("should create ted payment method", func(t *testing.T) {
		pm, err := vos.NewPaymentMethod("ted")
		require.NoError(t, err)
		require.False(t, pm.IsCredit())
		require.False(t, pm.RequiresCard())
		require.Equal(t, "ted", pm.String())
	})

	t.Run("should create debit payment method", func(t *testing.T) {
		pm, err := vos.NewPaymentMethod("debit")
		require.NoError(t, err)
		require.False(t, pm.IsCredit())
		require.True(t, pm.RequiresCard())
		require.Equal(t, "debit", pm.String())
	})

	t.Run("should create credit payment method", func(t *testing.T) {
		pm, err := vos.NewPaymentMethod("credit")
		require.NoError(t, err)
		require.True(t, pm.IsCredit())
		require.True(t, pm.RequiresCard())
		require.Equal(t, "credit", pm.String())
	})

	t.Run("should return error for uppercase Credit", func(t *testing.T) {
		_, err := vos.NewPaymentMethod("Credit")
		require.ErrorIs(t, err, domain.ErrInvalidPaymentMethod)
	})

	t.Run("should return error for unknown method wire", func(t *testing.T) {
		_, err := vos.NewPaymentMethod("wire")
		require.ErrorIs(t, err, domain.ErrInvalidPaymentMethod)
	})

	t.Run("should return error for empty string", func(t *testing.T) {
		_, err := vos.NewPaymentMethod("")
		require.ErrorIs(t, err, domain.ErrInvalidPaymentMethod)
	})
}
