package entities

import "errors"

var (
	// TransactionItem errors.
	ErrInvalidMonthlyID     = errors.New("invalid monthly transaction id")
	ErrInvalidCategoryID    = errors.New("invalid category id")
	ErrTitleRequired        = errors.New("title is required")
	ErrTitleTooLong         = errors.New("title is too long (max 255 characters)")
	ErrAmountMustBePositive = errors.New("amount must be positive")
	ErrInvalidDirection     = errors.New("invalid transaction direction")
	ErrInvalidType          = errors.New("invalid transaction type")

	// MonthlyTransaction errors.
	ErrInvalidUserID               = errors.New("invalid user id")
	ErrInvalidReferenceMonth       = errors.New("invalid reference month")
	ErrItemNotFound                = errors.New("transaction item not found")
	ErrItemDoesNotBelong           = errors.New("item does not belong to this monthly transaction")
	ErrCreditCardItemAlreadyExists = errors.New("credit card item already exists for this month")
)
