package factories_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	domain "github.com/jailtonjunior94/financial/internal/card/domain"
	"github.com/jailtonjunior94/financial/internal/card/domain/factories"
)

func TestCreateCard(t *testing.T) {
	t.Run("should create credit card with all fields", func(t *testing.T) {
		params := factories.CreateCardParams{
			UserID:            "550e8400-e29b-41d4-a716-446655440000",
			Name:              "Nubank Platinum",
			Type:              "credit",
			Flag:              "mastercard",
			LastFourDigits:    "1234",
			DueDay:            15,
			ClosingOffsetDays: 7,
		}

		card, err := factories.CreateCard(params)

		require.NoError(t, err)
		require.NotNil(t, card)
		require.NotEmpty(t, card.ID.String())
		require.Equal(t, "credit", card.Type.Value)
		require.Equal(t, "mastercard", card.Flag.Value)
		require.Equal(t, "1234", card.LastFourDigits.Value)
		require.Equal(t, 15, card.DueDay.Int())
		require.Equal(t, 7, card.ClosingOffsetDays.Int())
	})

	t.Run("should create debit card without due day", func(t *testing.T) {
		params := factories.CreateCardParams{
			UserID:         "550e8400-e29b-41d4-a716-446655440000",
			Name:           "Nubank Debito",
			Type:           "debit",
			Flag:           "visa",
			LastFourDigits: "5678",
		}

		card, err := factories.CreateCard(params)

		require.NoError(t, err)
		require.NotNil(t, card)
		require.Equal(t, "debit", card.Type.Value)
		require.Equal(t, 0, card.DueDay.Int())
		require.Equal(t, 0, card.ClosingOffsetDays.Int())
	})

	t.Run("should apply default closing offset days for credit without closing offset", func(t *testing.T) {
		params := factories.CreateCardParams{
			UserID:         "550e8400-e29b-41d4-a716-446655440000",
			Name:           "Inter",
			Type:           "credit",
			Flag:           "elo",
			LastFourDigits: "9999",
			DueDay:         10,
		}

		card, err := factories.CreateCard(params)

		require.NoError(t, err)
		require.Equal(t, 7, card.ClosingOffsetDays.Int())
	})

	t.Run("should return error for invalid user id", func(t *testing.T) {
		params := factories.CreateCardParams{
			UserID:         "invalid-uuid",
			Name:           "Nubank",
			Type:           "credit",
			Flag:           "visa",
			LastFourDigits: "1234",
			DueDay:         15,
		}

		_, err := factories.CreateCard(params)

		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid user_id")
	})

	t.Run("should return error for empty name", func(t *testing.T) {
		params := factories.CreateCardParams{
			UserID:         "550e8400-e29b-41d4-a716-446655440000",
			Name:           "",
			Type:           "credit",
			Flag:           "visa",
			LastFourDigits: "1234",
			DueDay:         15,
		}

		_, err := factories.CreateCard(params)

		require.Error(t, err)
	})

	t.Run("should return error for invalid type", func(t *testing.T) {
		params := factories.CreateCardParams{
			UserID:         "550e8400-e29b-41d4-a716-446655440000",
			Name:           "Nubank",
			Type:           "prepaid",
			Flag:           "visa",
			LastFourDigits: "1234",
			DueDay:         15,
		}

		_, err := factories.CreateCard(params)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidCardType)
	})

	t.Run("should return error for invalid flag", func(t *testing.T) {
		params := factories.CreateCardParams{
			UserID:         "550e8400-e29b-41d4-a716-446655440000",
			Name:           "Nubank",
			Type:           "credit",
			Flag:           "dinners",
			LastFourDigits: "1234",
			DueDay:         15,
		}

		_, err := factories.CreateCard(params)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidCardFlag)
	})

	t.Run("should return error for invalid last four digits", func(t *testing.T) {
		params := factories.CreateCardParams{
			UserID:         "550e8400-e29b-41d4-a716-446655440000",
			Name:           "Nubank",
			Type:           "credit",
			Flag:           "visa",
			LastFourDigits: "abcd",
			DueDay:         15,
		}

		_, err := factories.CreateCard(params)

		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrInvalidLastFourDigits)
	})
}
