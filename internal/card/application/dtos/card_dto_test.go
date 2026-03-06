package dtos_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financial/internal/card/application/dtos"
)

func intPtr(v int) *int {
	return &v
}

func TestCardInput_Validate(t *testing.T) {
	t.Run("should validate credit card with all fields", func(t *testing.T) {
		input := &dtos.CardInput{
			Name:           "Nubank",
			Type:           "credit",
			Flag:           "mastercard",
			LastFourDigits: "1234",
			DueDay:         intPtr(10),
		}
		errs := input.Validate()
		require.False(t, errs.HasErrors())
	})

	t.Run("should validate debit card without due_day", func(t *testing.T) {
		input := &dtos.CardInput{
			Name:           "Nubank Debito",
			Type:           "debit",
			Flag:           "visa",
			LastFourDigits: "5678",
		}
		errs := input.Validate()
		require.False(t, errs.HasErrors())
	})

	t.Run("should return error when credit card missing due_day", func(t *testing.T) {
		input := &dtos.CardInput{
			Name:           "Nubank",
			Type:           "credit",
			Flag:           "mastercard",
			LastFourDigits: "1234",
		}
		errs := input.Validate()
		require.True(t, errs.HasErrors())
	})

	t.Run("should return error when due_day is zero for credit", func(t *testing.T) {
		input := &dtos.CardInput{
			Name:           "Nubank",
			Type:           "credit",
			Flag:           "mastercard",
			LastFourDigits: "1234",
			DueDay:         intPtr(0),
		}
		errs := input.Validate()
		require.True(t, errs.HasErrors())
	})

	t.Run("should return error when due_day exceeds 31 for credit", func(t *testing.T) {
		input := &dtos.CardInput{
			Name:           "Nubank",
			Type:           "credit",
			Flag:           "mastercard",
			LastFourDigits: "1234",
			DueDay:         intPtr(32),
		}
		errs := input.Validate()
		require.True(t, errs.HasErrors())
	})

	t.Run("should return error when name is empty", func(t *testing.T) {
		input := &dtos.CardInput{
			Name:           "",
			Type:           "credit",
			Flag:           "mastercard",
			LastFourDigits: "1234",
			DueDay:         intPtr(10),
		}
		errs := input.Validate()
		require.True(t, errs.HasErrors())
	})

	t.Run("should return error when type is empty", func(t *testing.T) {
		input := &dtos.CardInput{
			Name:           "Nubank",
			Type:           "",
			Flag:           "mastercard",
			LastFourDigits: "1234",
		}
		errs := input.Validate()
		require.True(t, errs.HasErrors())
	})

	t.Run("should return error when type is invalid", func(t *testing.T) {
		input := &dtos.CardInput{
			Name:           "Nubank",
			Type:           "prepaid",
			Flag:           "mastercard",
			LastFourDigits: "1234",
		}
		errs := input.Validate()
		require.True(t, errs.HasErrors())
	})

	t.Run("should return error when flag is empty", func(t *testing.T) {
		input := &dtos.CardInput{
			Name:           "Nubank",
			Type:           "credit",
			Flag:           "",
			LastFourDigits: "1234",
			DueDay:         intPtr(10),
		}
		errs := input.Validate()
		require.True(t, errs.HasErrors())
	})

	t.Run("should return error when flag is invalid", func(t *testing.T) {
		input := &dtos.CardInput{
			Name:           "Nubank",
			Type:           "credit",
			Flag:           "dinners",
			LastFourDigits: "1234",
			DueDay:         intPtr(10),
		}
		errs := input.Validate()
		require.True(t, errs.HasErrors())
	})

	t.Run("should return error when last_four_digits is empty", func(t *testing.T) {
		input := &dtos.CardInput{
			Name:           "Nubank",
			Type:           "credit",
			Flag:           "mastercard",
			LastFourDigits: "",
			DueDay:         intPtr(10),
		}
		errs := input.Validate()
		require.True(t, errs.HasErrors())
	})

	t.Run("should return error when last_four_digits has non-numeric characters", func(t *testing.T) {
		input := &dtos.CardInput{
			Name:           "Nubank",
			Type:           "credit",
			Flag:           "mastercard",
			LastFourDigits: "12ab",
			DueDay:         intPtr(10),
		}
		errs := input.Validate()
		require.True(t, errs.HasErrors())
	})

	t.Run("should return error when last_four_digits has wrong length", func(t *testing.T) {
		input := &dtos.CardInput{
			Name:           "Nubank",
			Type:           "credit",
			Flag:           "mastercard",
			LastFourDigits: "123",
			DueDay:         intPtr(10),
		}
		errs := input.Validate()
		require.True(t, errs.HasErrors())
	})

	t.Run("should not return error when credit closing_offset_days is nil", func(t *testing.T) {
		input := &dtos.CardInput{
			Name:           "Nubank",
			Type:           "credit",
			Flag:           "mastercard",
			LastFourDigits: "1234",
			DueDay:         intPtr(10),
		}
		errs := input.Validate()
		require.False(t, errs.HasErrors())
	})

	t.Run("should return error when closing_offset_days exceeds 31", func(t *testing.T) {
		input := &dtos.CardInput{
			Name:              "Nubank",
			Type:              "credit",
			Flag:              "mastercard",
			LastFourDigits:    "1234",
			DueDay:            intPtr(10),
			ClosingOffsetDays: intPtr(35),
		}
		errs := input.Validate()
		require.True(t, errs.HasErrors())
	})

	t.Run("should ignore due_day for debit card", func(t *testing.T) {
		input := &dtos.CardInput{
			Name:           "Nubank Debito",
			Type:           "debit",
			Flag:           "visa",
			LastFourDigits: "5678",
			DueDay:         intPtr(10),
		}
		errs := input.Validate()
		require.False(t, errs.HasErrors())
	})
}

func TestCardUpdateInput_Validate(t *testing.T) {
	t.Run("should validate update input successfully", func(t *testing.T) {
		input := &dtos.CardUpdateInput{
			Name:           "Nubank Platinum",
			Flag:           "mastercard",
			LastFourDigits: "1234",
			DueDay:         intPtr(15),
		}
		errs := input.Validate()
		require.False(t, errs.HasErrors())
	})

	t.Run("should validate update input without due_day", func(t *testing.T) {
		input := &dtos.CardUpdateInput{
			Name:           "Nubank",
			Flag:           "visa",
			LastFourDigits: "5678",
		}
		errs := input.Validate()
		require.False(t, errs.HasErrors())
	})

	t.Run("should return error when name is empty", func(t *testing.T) {
		input := &dtos.CardUpdateInput{
			Name:           "",
			Flag:           "visa",
			LastFourDigits: "1234",
		}
		errs := input.Validate()
		require.True(t, errs.HasErrors())
	})

	t.Run("should return error when flag is invalid", func(t *testing.T) {
		input := &dtos.CardUpdateInput{
			Name:           "Nubank",
			Flag:           "dinners",
			LastFourDigits: "1234",
		}
		errs := input.Validate()
		require.True(t, errs.HasErrors())
	})

	t.Run("should return error when last_four_digits has non-numeric characters", func(t *testing.T) {
		input := &dtos.CardUpdateInput{
			Name:           "Nubank",
			Flag:           "visa",
			LastFourDigits: "12ab",
		}
		errs := input.Validate()
		require.True(t, errs.HasErrors())
	})

	t.Run("should return error when due_day is out of range", func(t *testing.T) {
		input := &dtos.CardUpdateInput{
			Name:           "Nubank",
			Flag:           "visa",
			LastFourDigits: "1234",
			DueDay:         intPtr(0),
		}
		errs := input.Validate()
		require.True(t, errs.HasErrors())
	})

	t.Run("should return error when closing_offset_days is out of range", func(t *testing.T) {
		input := &dtos.CardUpdateInput{
			Name:              "Nubank",
			Flag:              "visa",
			LastFourDigits:    "1234",
			ClosingOffsetDays: intPtr(50),
		}
		errs := input.Validate()
		require.True(t, errs.HasErrors())
	})
}
