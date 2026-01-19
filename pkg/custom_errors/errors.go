package customerrors

import "errors"

var (
	// Authentication errors.
	ErrUnauthorized         = errors.New("unauthorized: user not found in context")
	ErrMissingAuthHeader    = errors.New("missing authorization header")
	ErrInvalidAuthFormat    = errors.New("invalid authorization format: expected 'Bearer <token>'")
	ErrInvalidToken         = errors.New("invalid or malformed token")
	ErrTokenExpired         = errors.New("token has expired")
	ErrEmptyToken           = errors.New("empty token provided")
	ErrInvalidTokenClaims   = errors.New("invalid token claims")
	ErrInvalidSigningMethod = errors.New("invalid token signing method")

	// Domain errors.
	ErrBudgetNotFound        = errors.New("budget not found")
	ErrCardNotFound          = errors.New("card not found")
	ErrCategoryNotFound      = errors.New("category not found")
	ErrPaymentMethodNotFound = errors.New("payment method not found")
	ErrCannotBeEmpty         = errors.New("name cannot be empty")
	ErrInvalidEmail          = errors.New("invalid email format")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrNameCannotBeEmpty     = errors.New("name cannot be empty")
	ErrTooLong               = errors.New("name cannot be more than 255 characters")
	ErrNameTooLong           = errors.New("name cannot be more than 100 characters")
	ErrPasswordIsRequired    = errors.New("password is required")
	ErrUserNotFound          = errors.New("user not found")
	ErrCheckHash             = errors.New("error checking hash")
	ErrNameIsRequired        = errors.New("name is required")
	ErrUserIDIsRequired      = errors.New("user_id is required")
	ErrSequenceIsRequired    = errors.New("sequence is required")
	ErrSequenceTooLarge      = errors.New("sequence cannot be greater than 1000")
	ErrCategoryCycle         = errors.New("category cannot be its own parent or create a cycle")
	ErrBudgetInvalidTotal    = errors.New("sum of budget item percentages must equal 100%")
	ErrInvalidParentCategory = errors.New("parent category not found or belongs to different user")

	// Validation errors (shared across modules).
	ErrNegativeAmount    = errors.New("amount cannot be negative")
	ErrInvalidCategoryID = errors.New("invalid category ID")
)
