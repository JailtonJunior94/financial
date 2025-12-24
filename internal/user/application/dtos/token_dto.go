package dtos

import "time"

type AuthInput struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

type AuthOutput struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewAuthOutput(token string, authExpirationAt int) *AuthOutput {
	return &AuthOutput{
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour * time.Duration(authExpirationAt)),
	}
}
