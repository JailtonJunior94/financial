package domain

import "errors"

var (
	// Invoice errors
	ErrInvoiceNotFound              = errors.New("invoice not found")
	ErrInvoiceAlreadyExistsForMonth = errors.New("invoice already exists for this card and month")
	ErrInvoiceHasNoItems            = errors.New("invoice must have at least one item")
	ErrInvoiceNegativeTotal         = errors.New("invoice total amount cannot be negative")

	// InvoiceItem errors
	ErrInvoiceItemNotFound      = errors.New("invoice item not found")
	ErrInvalidPurchaseDate      = errors.New("purchase date cannot be in the future")
	ErrNegativeAmount           = errors.New("amount cannot be negative")
	ErrInvalidInstallment       = errors.New("installment number must be between 1 and installment total")
	ErrInvalidInstallmentTotal  = errors.New("installment total must be at least 1")
	ErrInstallmentAmountInvalid = errors.New("installment amount must equal total amount divided by installments")
	ErrInvalidCategoryID        = errors.New("invalid category ID")
	ErrInvalidCardID            = errors.New("invalid card ID")
	ErrEmptyDescription         = errors.New("description cannot be empty")
)
