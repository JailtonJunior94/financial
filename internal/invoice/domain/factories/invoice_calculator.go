package factories

import (
	"fmt"
	"time"

	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

// InvoiceCalculator implements per-instance billing cycle calculation.
//
// Each card can have different due day and closing offset days.
// closing_day = due_day - closing_offset_days
//
// Allocation rule:
//
//	purchase_day <= closing_day → reference month = current month
//	purchase_day >  closing_day → reference month = next month
type InvoiceCalculator struct {
	dueDay            int
	closingOffsetDays int
}

// NewInvoiceCalculator creates an InvoiceCalculator for a specific card billing config.
//
// Constraints:
//   - dueDay must be in [1, 31]
//   - closingOffsetDays must be in [1, 31]
//   - dueDay must be greater than closingOffsetDays
func NewInvoiceCalculator(dueDay, closingOffsetDays int) (*InvoiceCalculator, error) {
	if dueDay < 1 || dueDay > 31 {
		return nil, fmt.Errorf("dueDay must be between 1 and 31, got %d", dueDay)
	}
	if closingOffsetDays < 1 || closingOffsetDays > 31 {
		return nil, fmt.Errorf("closingOffsetDays must be between 1 and 31, got %d", closingOffsetDays)
	}
	if dueDay <= closingOffsetDays {
		return nil, fmt.Errorf("dueDay (%d) must be greater than closingOffsetDays (%d)", dueDay, closingOffsetDays)
	}
	return &InvoiceCalculator{dueDay: dueDay, closingOffsetDays: closingOffsetDays}, nil
}

// ClosingDay returns the closing day of the billing cycle (dueDay - closingOffsetDays).
func (c *InvoiceCalculator) ClosingDay() int {
	return c.dueDay - c.closingOffsetDays
}

// CalculateInvoiceMonth determines the reference month for a purchase.
//
//	purchase_day <= closing_day → current month
//	purchase_day >  closing_day → next month
func (c *InvoiceCalculator) CalculateInvoiceMonth(purchaseDate time.Time) pkgVos.ReferenceMonth {
	closingDay := c.ClosingDay()

	y := purchaseDate.Year()
	m := int(purchaseDate.Month())

	offset := 0
	if purchaseDate.Day() > closingDay {
		offset = 1
	}

	totalMonths := y*12 + (m - 1) + offset
	refYear := totalMonths / 12
	refMonth := time.Month(totalMonths%12 + 1)

	return pkgVos.NewReferenceMonthFromDate(
		time.Date(refYear, refMonth, 1, 0, 0, 0, 0, time.UTC),
	)
}

// CalculateInstallmentMonths returns the reference months for each installment.
//
// The first installment uses CalculateInvoiceMonth; subsequent ones are consecutive months.
func (c *InvoiceCalculator) CalculateInstallmentMonths(purchaseDate time.Time, installmentTotal int) []pkgVos.ReferenceMonth {
	months := make([]pkgVos.ReferenceMonth, installmentTotal)
	firstMonth := c.CalculateInvoiceMonth(purchaseDate)
	months[0] = firstMonth
	for i := 1; i < installmentTotal; i++ {
		months[i] = firstMonth.AddMonths(i)
	}
	return months
}

// CalculateDueDate returns the due date for a given reference month.
func (c *InvoiceCalculator) CalculateDueDate(referenceMonth pkgVos.ReferenceMonth) time.Time {
	return time.Date(referenceMonth.Year(), referenceMonth.Month(), c.dueDay, 0, 0, 0, 0, time.UTC)
}
