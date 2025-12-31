package vos

import (
	"errors"
	"fmt"
)

var (
	ErrDueDayIsRequired = errors.New("due day is required")
	ErrDueDayInvalid    = errors.New("due day must be between 1 and 31")
)

type DueDay struct {
	Value int
	Valid bool
}

func NewDueDay(day int) (DueDay, error) {
	if day < 1 || day > 31 {
		return DueDay{}, fmt.Errorf("invalid due day: %w", ErrDueDayInvalid)
	}
	return DueDay{Value: day, Valid: true}, nil
}

func (v DueDay) Int() int {
	return v.Value
}
