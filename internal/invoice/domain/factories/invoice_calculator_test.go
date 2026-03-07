package factories_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financial/internal/invoice/domain/factories"
)

func date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}


func TestInvoiceCalculator(t *testing.T) {
	t.Run("should create calculator and return correct closing day", func(t *testing.T) {
		calc, err := factories.NewInvoiceCalculator(10, 7)
		require.NoError(t, err)
		require.Equal(t, 3, calc.ClosingDay())
	})

	t.Run("should return error when dueDay equals closingOffsetDays", func(t *testing.T) {
		_, err := factories.NewInvoiceCalculator(1, 1)
		require.Error(t, err)
	})

	t.Run("should return error when dueDay is zero", func(t *testing.T) {
		_, err := factories.NewInvoiceCalculator(0, 5)
		require.Error(t, err)
	})

	t.Run("should return error when closingOffsetDays exceeds 31", func(t *testing.T) {
		_, err := factories.NewInvoiceCalculator(10, 32)
		require.Error(t, err)
	})

	t.Run("should allocate to current month when purchase is before closing day", func(t *testing.T) {
		calc, err := factories.NewInvoiceCalculator(10, 7)
		require.NoError(t, err)
		purchaseDate := date(2026, 3, 2)
		month := calc.CalculateInvoiceMonth(purchaseDate)
		require.Equal(t, "2026-03", month.String())
	})

	t.Run("should allocate to current month when purchase is on closing day", func(t *testing.T) {
		calc, err := factories.NewInvoiceCalculator(10, 7)
		require.NoError(t, err)
		purchaseDate := date(2026, 3, 3)
		month := calc.CalculateInvoiceMonth(purchaseDate)
		require.Equal(t, "2026-03", month.String())
	})

	t.Run("should allocate to next month when purchase is after closing day", func(t *testing.T) {
		calc, err := factories.NewInvoiceCalculator(10, 7)
		require.NoError(t, err)
		purchaseDate := date(2026, 3, 4)
		month := calc.CalculateInvoiceMonth(purchaseDate)
		require.Equal(t, "2026-04", month.String())
	})

	t.Run("should return january of next year for december purchase after closing", func(t *testing.T) {
		calc, err := factories.NewInvoiceCalculator(10, 7)
		require.NoError(t, err)
		purchaseDate := date(2026, 12, 5)
		month := calc.CalculateInvoiceMonth(purchaseDate)
		require.Equal(t, "2027-01", month.String())
	})

	t.Run("should return january for january purchase before closing day", func(t *testing.T) {
		calc, err := factories.NewInvoiceCalculator(10, 7)
		require.NoError(t, err)
		purchaseDate := date(2026, 1, 1)
		month := calc.CalculateInvoiceMonth(purchaseDate)
		require.Equal(t, "2026-01", month.String())
	})

	t.Run("should return three consecutive months for three installments in march", func(t *testing.T) {
		calc, err := factories.NewInvoiceCalculator(10, 7)
		require.NoError(t, err)
		purchaseDate := date(2026, 3, 2)
		months := calc.CalculateInstallmentMonths(purchaseDate, 3)
		require.Len(t, months, 3)
		require.Equal(t, "2026-03", months[0].String())
		require.Equal(t, "2026-04", months[1].String())
		require.Equal(t, "2026-05", months[2].String())
	})

	t.Run("should return january and february for december purchase after closing", func(t *testing.T) {
		calc, err := factories.NewInvoiceCalculator(10, 7)
		require.NoError(t, err)
		purchaseDate := date(2026, 12, 5)
		months := calc.CalculateInstallmentMonths(purchaseDate, 2)
		require.Len(t, months, 2)
		require.Equal(t, "2027-01", months[0].String())
		require.Equal(t, "2027-02", months[1].String())
	})
}
