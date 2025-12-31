package domain

import "errors"

var (
	// Budget errors
	ErrBudgetNotFound              = errors.New("budget not found")
	ErrBudgetAlreadyExistsForMonth = errors.New("budget already exists for this month")
	ErrBudgetInvalidTotal          = errors.New("sum of budget item percentages must equal 100%")
	ErrBudgetPercentageExceeds100  = errors.New("sum of budget item percentages exceeds 100%")
	ErrBudgetNoItems               = errors.New("budget must have at least one item")

	// BudgetItem errors
	ErrBudgetItemNotFound = errors.New("budget item not found")
	ErrInvalidPercentage  = errors.New("percentage must be between 0 and 100")
	ErrNegativeAmount     = errors.New("amount cannot be negative")
	ErrInvalidCategoryID  = errors.New("invalid category ID")
	ErrDuplicateCategory  = errors.New("category already exists in budget")
)
