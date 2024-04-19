package vos

import (
	"errors"
	"strings"
)

var (
	ErrNameCannotBeEmpty = errors.New("name cannot be empty")
	ErrNameTooLong       = errors.New("name cannot be more than 100 characters")
)

type UserName struct {
	Value string
}

func NewUserName(value string) (UserName, error) {
	if len(strings.TrimSpace(value)) == 0 {
		return UserName{}, ErrCannotBeEmpty
	}

	if len(value) > 100 {
		return UserName{}, ErrTooLong
	}
	return UserName{Value: value}, nil
}

func (n UserName) String() string {
	return n.Value
}
