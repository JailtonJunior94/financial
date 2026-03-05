package auth

import "context"

// TokenGenerator defines the interface for token generation.
// Allows decoupling use cases that generate tokens from validation-only dependencies,
// following the Interface Segregation Principle.
type TokenGenerator interface {
	// GenerateToken creates a signed token for the given user identity.
	GenerateToken(ctx context.Context, id, email string) (string, error)
}
