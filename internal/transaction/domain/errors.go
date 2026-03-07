package domain

import "errors"

var (
	ErrTransactionNotFound       = errors.New("transaction not found")
	ErrTransactionNotOwned       = errors.New("transaction does not belong to user")
	ErrInvoiceClosed             = errors.New("invoice is closed and cannot be modified")
	ErrNothingToReverse          = errors.New("all installments are in closed or paid invoices")
	ErrCardRequiredForCredit     = errors.New("card_id is required for credit and debit payments")
	ErrCardNotAllowedForMethod   = errors.New("card_id is not allowed for this payment method")
	ErrInstallmentsOnlyForCredit = errors.New("installments only allowed for credit payment method")
	ErrTransactionDateFuture     = errors.New("transaction_date cannot be in the future")
	ErrInvalidPaymentMethod      = errors.New("invalid payment method")
	ErrInvalidTransactionStatus  = errors.New("invalid transaction status")
	ErrDescriptionRequired       = errors.New("description is required")
	ErrAmountMustBePositive      = errors.New("amount must be positive")
	ErrInstallmentsTooMany       = errors.New("installments cannot exceed 48")
)
