package error

import "errors"

var (
	ErrBudgetNotFound     = errors.New("budget not found")
	ErrCategoryNotFound   = errors.New("category not found")
	ErrCannotBeEmpty      = errors.New("name cannot be empty")
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrNameCannotBeEmpty  = errors.New("name cannot be empty")
	ErrTooLong            = errors.New("name cannot be more than 255 characters")
	ErrNameTooLong        = errors.New("name cannot be more than 100 characters")
	ErrPasswordIsRequired = errors.New("password is required")
	ErrUserNotFound       = errors.New("user not found")
	ErrCheckHash          = errors.New("error checking hash")
	ErrNameIsRequired     = errors.New("name is required")
	ErrUserIDIsRequired   = errors.New("user_id is required")
	ErrSequenceIsRequired = errors.New("sequence is required")
)
