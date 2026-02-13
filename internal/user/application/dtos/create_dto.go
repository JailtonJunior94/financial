package dtos

import "time"

type CreateUserInput struct {
	Name     string `json:"name"     example:"João Silva"`
	Email    string `json:"email"    example:"joao@email.com"`
	Password string `json:"password" example:"senhaSegura123"`
}

type CreateUserOutput struct {
	ID        string    `json:"id"         example:"550e8400-e29b-41d4-a716-446655440000"`
	Name      string    `json:"name"       example:"João Silva"`
	Email     string    `json:"email"      example:"joao@email.com"`
	CreatedAt time.Time `json:"created_at" example:"2025-01-15T10:30:00Z"`
}
