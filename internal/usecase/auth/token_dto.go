package auth

import "time"

type AuthInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthOutput struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
