package vos

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidTransactionType = errors.New("invalid transaction type")
)

// TransactionType representa o tipo de transação.
type TransactionType string

const (
	TypePix        TransactionType = "PIX"
	TypeBoleto     TransactionType = "BOLETO"
	TypeTransfer   TransactionType = "TRANSFER"
	TypeCreditCard TransactionType = "CREDIT_CARD"
)

// NewTransactionType cria um novo TransactionType validado.
func NewTransactionType(value string) (TransactionType, error) {
	transactionType := TransactionType(value)
	if !transactionType.IsValid() {
		return "", fmt.Errorf("%w: %s", ErrInvalidTransactionType, value)
	}
	return transactionType, nil
}

// IsValid verifica se o tipo é válido.
func (t TransactionType) IsValid() bool {
	switch t {
	case TypePix, TypeBoleto, TypeTransfer, TypeCreditCard:
		return true
	default:
		return false
	}
}

// String retorna a representação em string.
func (t TransactionType) String() string {
	return string(t)
}

// IsCreditCard verifica se é uma transação de cartão de crédito.
func (t TransactionType) IsCreditCard() bool {
	return t == TypeCreditCard
}
