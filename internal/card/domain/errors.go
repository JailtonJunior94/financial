package domain

import "errors"

var (
	ErrCardNotFound          = errors.New("card not found")
	ErrCardHasOpenInvoices   = errors.New("card has open invoices")
	ErrInvalidCardType       = errors.New("invalid card type: must be 'credit' or 'debit'")
	ErrInvalidCardFlag       = errors.New("invalid card flag")
	ErrInvalidLastFourDigits = errors.New("invalid last four digits: must be exactly 4 numeric digits")
	ErrDueDayRequired        = errors.New("due_day is required for credit cards")
)
