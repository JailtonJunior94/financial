package entities_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/card/domain/entities"
	"github.com/jailtonjunior94/financial/internal/card/domain/vos"
)

func TestNewCard(t *testing.T) {
	t.Run("should create credit card with all fields", func(t *testing.T) {
		userID := createUUID(t)
		name := createCardName(t, "Nubank")
		cardType, _ := vos.NewCardType("credit")
		flag, _ := vos.NewCardFlag("visa")
		digits, _ := vos.NewLastFourDigits("1234")
		dueDay, _ := vos.NewDueDay(15)
		offset, _ := vos.NewClosingOffsetDays(7)

		card, err := entities.NewCard(userID, name, cardType, flag, digits, dueDay, offset)

		require.NoError(t, err)
		require.NotNil(t, card)
		require.Equal(t, "credit", card.Type.Value)
		require.Equal(t, "visa", card.Flag.Value)
		require.Equal(t, "1234", card.LastFourDigits.Value)
		require.Equal(t, 15, card.DueDay.Int())
		require.Equal(t, 7, card.ClosingOffsetDays.Int())
		require.False(t, card.CreatedAt.ValueOr(time.Time{}).IsZero())
	})

	t.Run("should create debit card without due day and closing offset", func(t *testing.T) {
		userID := createUUID(t)
		name := createCardName(t, "Nubank Debito")
		cardType, _ := vos.NewCardType("debit")
		flag, _ := vos.NewCardFlag("mastercard")
		digits, _ := vos.NewLastFourDigits("5678")

		card, err := entities.NewCard(userID, name, cardType, flag, digits, vos.DueDay{}, vos.ClosingOffsetDays{})

		require.NoError(t, err)
		require.NotNil(t, card)
		require.Equal(t, "debit", card.Type.Value)
		require.Equal(t, 0, card.DueDay.Int())
		require.Equal(t, 0, card.ClosingOffsetDays.Int())
	})
}

func TestCardUpdate(t *testing.T) {
	t.Run("should update credit card fields", func(t *testing.T) {
		card := createCreditCard(t)

		err := card.Update("Updated Name", "amex", "9999", 20, 10)

		require.NoError(t, err)
		require.Equal(t, "Updated Name", card.Name.String())
		require.Equal(t, "amex", card.Flag.Value)
		require.Equal(t, "9999", card.LastFourDigits.Value)
		require.Equal(t, 20, card.DueDay.Int())
		require.Equal(t, 10, card.ClosingOffsetDays.Int())
		require.False(t, card.UpdatedAt.ValueOr(time.Time{}).IsZero())
	})

	t.Run("should update debit card and ignore due day and closing offset", func(t *testing.T) {
		card := createDebitCard(t)

		err := card.Update("Updated Debit", "elo", "4321", 15, 7)

		require.NoError(t, err)
		require.Equal(t, "Updated Debit", card.Name.String())
		require.Equal(t, "elo", card.Flag.Value)
		require.Equal(t, "4321", card.LastFourDigits.Value)
		require.Equal(t, 0, card.DueDay.Int())
		require.Equal(t, 0, card.ClosingOffsetDays.Int())
	})

	t.Run("should return error for empty name", func(t *testing.T) {
		card := createCreditCard(t)

		err := card.Update("", "visa", "1234", 15, 7)

		require.Error(t, err)
	})

	t.Run("should return error for invalid flag", func(t *testing.T) {
		card := createCreditCard(t)

		err := card.Update("Valid Name", "invalid", "1234", 15, 7)

		require.Error(t, err)
	})

	t.Run("should return error for invalid last four digits", func(t *testing.T) {
		card := createCreditCard(t)

		err := card.Update("Valid Name", "visa", "abcd", 15, 7)

		require.Error(t, err)
	})
}

func TestCardDelete(t *testing.T) {
	t.Run("should soft delete card", func(t *testing.T) {
		card := createCreditCard(t)

		result := card.Delete()

		require.NotNil(t, result)
		require.False(t, result.DeletedAt.ValueOr(time.Time{}).IsZero())
		require.Equal(t, card, result)
	})
}

func createUUID(t *testing.T) sharedVos.UUID {
	t.Helper()
	uuid, err := sharedVos.NewUUID()
	require.NoError(t, err)
	return uuid
}

func createCardName(t *testing.T, name string) vos.CardName {
	t.Helper()
	cardName, err := vos.NewCardName(name)
	require.NoError(t, err)
	return cardName
}

func createCreditCard(t *testing.T) *entities.Card {
	t.Helper()
	userID := createUUID(t)
	name := createCardName(t, "Test Credit Card")
	cardType, _ := vos.NewCardType("credit")
	flag, _ := vos.NewCardFlag("visa")
	digits, _ := vos.NewLastFourDigits("1234")
	dueDay, _ := vos.NewDueDay(15)
	offset, _ := vos.NewClosingOffsetDays(7)
	card, err := entities.NewCard(userID, name, cardType, flag, digits, dueDay, offset)
	require.NoError(t, err)
	return card
}

func createDebitCard(t *testing.T) *entities.Card {
	t.Helper()
	userID := createUUID(t)
	name := createCardName(t, "Test Debit Card")
	cardType, _ := vos.NewCardType("debit")
	flag, _ := vos.NewCardFlag("mastercard")
	digits, _ := vos.NewLastFourDigits("5678")
	card, err := entities.NewCard(userID, name, cardType, flag, digits, vos.DueDay{}, vos.ClosingOffsetDays{})
	require.NoError(t, err)
	return card
}
