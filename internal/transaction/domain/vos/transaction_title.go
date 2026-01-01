package vos

import (
	"strings"

	"github.com/jailtonjunior94/financial/internal/transaction/domain"
)

type TransactionTitle struct {
	value string
}

func NewTransactionTitle(title string) (TransactionTitle, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return TransactionTitle{}, domain.ErrInvalidTransactionTitle
	}

	if len(title) < 3 {
		return TransactionTitle{}, domain.ErrInvalidTransactionTitle
	}

	if len(title) > 100 {
		return TransactionTitle{}, domain.ErrInvalidTransactionTitle
	}

	return TransactionTitle{value: title}, nil
}

func (t TransactionTitle) String() string {
	return t.value
}
