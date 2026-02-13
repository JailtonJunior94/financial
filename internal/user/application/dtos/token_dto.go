package dtos

import "time"

type AuthInput struct {
	Email    string `json:"email"    form:"email"    example:"joao@email.com"`
	Password string `json:"password" form:"password" example:"senhaSegura123"`
}

type AuthOutput struct {
	Token     string    `json:"token"      example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt time.Time `json:"expires_at" example:"2025-01-16T10:30:00Z"`
}

func NewAuthOutput(token string, authExpirationAt int) *AuthOutput {
	return &AuthOutput{
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour * time.Duration(authExpirationAt)),
	}
}
