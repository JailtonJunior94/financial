package vos

import (
	"net/mail"
	"strings"

	financialError "github.com/jailtonjunior94/financial/pkg/error"
)

type Email struct {
	Value string
}

func NewEmail(value string) (Email, error) {
	if len(strings.TrimSpace(value)) == 0 {
		return Email{}, financialError.ErrCannotBeEmpty
	}

	if len(value) > 255 {
		return Email{}, financialError.ErrTooLong
	}

	if _, err := mail.ParseAddress(value); err != nil {
		return Email{}, financialError.ErrInvalidEmail
	}
	return Email{Value: value}, nil
}

func (n Email) String() string {
	return n.Value
}
