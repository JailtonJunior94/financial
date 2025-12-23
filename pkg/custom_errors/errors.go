package customerrors

import "errors"

var (
	ErrUnauthorized           = errors.New("unauthorized: user not found in context")
	ErrBudgetNotFound         = errors.New("budget not found")
	ErrCategoryNotFound       = errors.New("category not found")
	ErrCannotBeEmpty          = errors.New("name cannot be empty")
	ErrInvalidEmail           = errors.New("invalid email format")
	ErrEmailAlreadyExists     = errors.New("email already exists")
	ErrNameCannotBeEmpty      = errors.New("name cannot be empty")
	ErrTooLong                = errors.New("name cannot be more than 255 characters")
	ErrNameTooLong            = errors.New("name cannot be more than 100 characters")
	ErrPasswordIsRequired     = errors.New("password is required")
	ErrUserNotFound           = errors.New("user not found")
	ErrCheckHash              = errors.New("error checking hash")
	ErrNameIsRequired         = errors.New("name is required")
	ErrUserIDIsRequired       = errors.New("user_id is required")
	ErrSequenceIsRequired     = errors.New("sequence is required")
	ErrSequenceTooLarge       = errors.New("sequence cannot be greater than 1000")
	ErrCategoryCycle          = errors.New("category cannot be its own parent or create a cycle")
	ErrBudgetInvalidTotal     = errors.New("sum of budget item percentages must equal 100%")
	ErrInvalidParentCategory  = errors.New("parent category not found or belongs to different user")
)
