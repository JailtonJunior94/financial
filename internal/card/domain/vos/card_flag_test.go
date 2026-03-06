package vos_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	domain "github.com/jailtonjunior94/financial/internal/card/domain"
	"github.com/jailtonjunior94/financial/internal/card/domain/vos"
)

func TestNewCardFlag(t *testing.T) {
	t.Run("should create valid visa flag", func(t *testing.T) {
		cf, err := vos.NewCardFlag("visa")
		require.NoError(t, err)
		require.Equal(t, "visa", cf.String())
	})

	t.Run("should create valid mastercard flag", func(t *testing.T) {
		cf, err := vos.NewCardFlag("mastercard")
		require.NoError(t, err)
		require.Equal(t, "mastercard", cf.String())
	})

	t.Run("should create valid elo flag", func(t *testing.T) {
		cf, err := vos.NewCardFlag("elo")
		require.NoError(t, err)
		require.Equal(t, "elo", cf.String())
	})

	t.Run("should create valid amex flag", func(t *testing.T) {
		cf, err := vos.NewCardFlag("amex")
		require.NoError(t, err)
		require.Equal(t, "amex", cf.String())
	})

	t.Run("should create valid hipercard flag", func(t *testing.T) {
		cf, err := vos.NewCardFlag("hipercard")
		require.NoError(t, err)
		require.Equal(t, "hipercard", cf.String())
	})

	t.Run("should return error for invalid flag", func(t *testing.T) {
		_, err := vos.NewCardFlag("dinners")
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidCardFlag)
	})

	t.Run("should return error for empty flag", func(t *testing.T) {
		_, err := vos.NewCardFlag("")
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidCardFlag)
	})
}
