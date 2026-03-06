package vos_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	domain "github.com/jailtonjunior94/financial/internal/card/domain"
	"github.com/jailtonjunior94/financial/internal/card/domain/vos"
)

func TestNewLastFourDigits(t *testing.T) {
	t.Run("should create valid last four digits", func(t *testing.T) {
		l, err := vos.NewLastFourDigits("1234")
		require.NoError(t, err)
		require.Equal(t, "1234", l.String())
	})

	t.Run("should create valid all zeros", func(t *testing.T) {
		l, err := vos.NewLastFourDigits("0000")
		require.NoError(t, err)
		require.Equal(t, "0000", l.String())
	})

	t.Run("should return error for letters", func(t *testing.T) {
		_, err := vos.NewLastFourDigits("abcd")
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidLastFourDigits)
	})

	t.Run("should return error for less than 4 digits", func(t *testing.T) {
		_, err := vos.NewLastFourDigits("123")
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidLastFourDigits)
	})

	t.Run("should return error for more than 4 digits", func(t *testing.T) {
		_, err := vos.NewLastFourDigits("12345")
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidLastFourDigits)
	})

	t.Run("should return error for empty string", func(t *testing.T) {
		_, err := vos.NewLastFourDigits("")
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidLastFourDigits)
	})

	t.Run("should return error for mixed digits and letters", func(t *testing.T) {
		_, err := vos.NewLastFourDigits("12a4")
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidLastFourDigits)
	})
}
