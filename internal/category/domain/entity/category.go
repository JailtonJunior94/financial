package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNameIsRequired     = errors.New("name is required")
	ErrUserIDIsRequired   = errors.New("user_id is required")
	ErrSequenceIsRequired = errors.New("sequence is required")
)

type Category struct {
	ID        string
	UserID    string
	Name      string
	Sequence  int
	CreatedAt time.Time
	UpdatedAt *time.Time
	Active    bool
}

func NewCategory(userID, name string, sequence int) (*Category, error) {
	category := &Category{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      name,
		Sequence:  sequence,
		CreatedAt: time.Now(),
		Active:    true,
	}
	if err := category.IsValid(); err != nil {
		return nil, err
	}
	return category, nil
}

func (u *Category) IsValid() error {
	if u.UserID == "" {
		return ErrUserIDIsRequired
	}
	if u.Name == "" {
		return ErrNameIsRequired
	}
	if u.Sequence < 0 {
		return ErrSequenceIsRequired
	}
	return nil
}
