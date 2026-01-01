package vos

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidTransactionDirection = errors.New("invalid transaction direction")
)

// TransactionDirection representa a direção de uma transação (entrada ou saída).
type TransactionDirection string

const (
	DirectionIncome  TransactionDirection = "INCOME"  // Entrada de dinheiro
	DirectionExpense TransactionDirection = "EXPENSE" // Saída de dinheiro
)

// NewTransactionDirection cria um novo TransactionDirection validado.
func NewTransactionDirection(value string) (TransactionDirection, error) {
	direction := TransactionDirection(value)
	if !direction.IsValid() {
		return "", fmt.Errorf("%w: %s", ErrInvalidTransactionDirection, value)
	}
	return direction, nil
}

// IsValid verifica se a direção é válida.
func (d TransactionDirection) IsValid() bool {
	switch d {
	case DirectionIncome, DirectionExpense:
		return true
	default:
		return false
	}
}

// String retorna a representação em string.
func (d TransactionDirection) String() string {
	return string(d)
}

// IsIncome verifica se é uma entrada.
func (d TransactionDirection) IsIncome() bool {
	return d == DirectionIncome
}

// IsExpense verifica se é uma saída.
func (d TransactionDirection) IsExpense() bool {
	return d == DirectionExpense
}
