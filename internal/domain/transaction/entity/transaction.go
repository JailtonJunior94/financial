package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrDescriptionIsRequired = errors.New("description is required")
)

type Transaction struct {
	ID        string
	UserID    string
	Date      time.Time
	Income    float64
	Outcome   float64
	CreatedAt time.Time
	UpdatedAt *time.Time
	Active    bool
}

func NewUser(userID string) (*Transaction, error) {
	transaction := &Transaction{
		ID:        uuid.New().String(),
		UserID:    userID,
		CreatedAt: time.Now(),
		Active:    true,
	}
	return transaction, nil
}
