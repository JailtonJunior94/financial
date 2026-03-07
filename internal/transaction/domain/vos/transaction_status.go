package vos

import "github.com/jailtonjunior94/financial/internal/transaction/domain"

const (
	TransactionStatusActive    = "active"
	TransactionStatusCancelled = "cancelled"
)

type TransactionStatus struct {
	Value string
}

func NewTransactionStatus(v string) (TransactionStatus, error) {
	switch v {
	case TransactionStatusActive, TransactionStatusCancelled:
		return TransactionStatus{Value: v}, nil
	default:
		return TransactionStatus{}, domain.ErrInvalidTransactionStatus
	}
}

func (s TransactionStatus) IsActive() bool {
	return s.Value == TransactionStatusActive
}

func (s TransactionStatus) IsCancelled() bool {
	return s.Value == TransactionStatusCancelled
}

func (s TransactionStatus) String() string {
	return s.Value
}
