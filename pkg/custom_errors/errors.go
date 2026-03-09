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

	// Domain errors (cross-cutting — not owned by a single module).
	ErrCannotBeEmpty         = errors.New("name cannot be empty")
	ErrInvalidEmail          = errors.New("invalid email format")
	ErrTooLong               = errors.New("name cannot be more than 255 characters")
	ErrNameTooLong           = errors.New("name cannot be more than 100 characters")
	ErrPasswordIsRequired    = errors.New("password is required")
	ErrCheckHash             = errors.New("error checking hash")
	ErrNameIsRequired        = errors.New("name is required")
	ErrUserIDIsRequired      = errors.New("user_id is required")
	ErrSequenceIsRequired    = errors.New("sequence is required")
	ErrSequenceTooLarge      = errors.New("sequence cannot be greater than 1000")
	ErrCategoryCycle         = errors.New("category cannot be its own parent or create a cycle")
	ErrInvalidParentCategory = errors.New("parent category not found or belongs to different user")
	ErrSubcategoryNotFound   = errors.New("subcategory not found")

	// Authorization errors.
	ErrForbidden = errors.New("you do not have permission to access this resource")
)
