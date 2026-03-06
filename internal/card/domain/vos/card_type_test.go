package vos_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	domain "github.com/jailtonjunior94/financial/internal/card/domain"
	"github.com/jailtonjunior94/financial/internal/card/domain/vos"
)

func TestNewCardType(t *testing.T) {
	t.Run("should create valid credit type", func(t *testing.T) {
		ct, err := vos.NewCardType("credit")
		require.NoError(t, err)
		require.True(t, ct.IsCredit())
		require.False(t, ct.IsDebit())
		require.Equal(t, "credit", ct.String())
	})

	t.Run("should create valid debit type", func(t *testing.T) {
		ct, err := vos.NewCardType("debit")
		require.NoError(t, err)
		require.True(t, ct.IsDebit())
		require.False(t, ct.IsCredit())
		require.Equal(t, "debit", ct.String())
	})

	t.Run("should return error for invalid type", func(t *testing.T) {
		_, err := vos.NewCardType("invalid")
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidCardType)
	})

	t.Run("should return error for empty type", func(t *testing.T) {
		_, err := vos.NewCardType("")
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidCardType)
	})

	t.Run("should return error for case-sensitive mismatch", func(t *testing.T) {
		_, err := vos.NewCardType("Credit")
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidCardType)
	})
}
