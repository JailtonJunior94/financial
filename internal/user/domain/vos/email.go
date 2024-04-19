package vos

import (
	"errors"
	"net/mail"
	"strings"
)

var (
	ErrCannotBeEmpty = errors.New("name cannot be empty")
	ErrInvalidEmail  = errors.New("invalid email format")
	ErrTooLong       = errors.New("name cannot be more than 255 characters")
)

type Email struct {
	Value string
}

func NewEmail(value string) (Email, error) {
	if len(strings.TrimSpace(value)) == 0 {
		return Email{}, ErrCannotBeEmpty
	}

	if len(value) > 255 {
		return Email{}, ErrTooLong
	}

	if _, err := mail.ParseAddress(value); err != nil {
		return Email{}, ErrInvalidEmail
	}
	return Email{Value: value}, nil
}

func (n Email) String() string {
	return n.Value
}
