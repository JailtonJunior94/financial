package vos

import (
	"strings"

	financialError "github.com/jailtonjunior94/financial/pkg/custom_errors"
)

var ()

type UserName struct {
	Value string
}

func NewUserName(value string) (UserName, error) {
	if len(strings.TrimSpace(value)) == 0 {
		return UserName{}, financialError.ErrCannotBeEmpty
	}

	if len(value) > 100 {
		return UserName{}, financialError.ErrTooLong
	}
	return UserName{Value: value}, nil
}

func (n UserName) String() string {
	return n.Value
}
