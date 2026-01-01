package domain

import "errors"

var (
	// Erros de MonthlyTransaction
	ErrMonthlyTransactionNotFound      = errors.New("transação mensal não encontrada")
	ErrMonthlyTransactionAlreadyExists = errors.New("transação mensal já existe para este mês")
	ErrInvalidReferenceMonth           = errors.New("mês de referência inválido")

	// Erros de TransactionItem
	ErrTransactionItemNotFound     = errors.New("item de transação não encontrado")
	ErrTransactionItemDeleted      = errors.New("item de transação foi excluído")
	ErrInvalidTransactionTitle     = errors.New("título da transação inválido")
	ErrInvalidTransactionAmount    = errors.New("valor da transação inválido")
	ErrInvalidTransactionDirection = errors.New("direção da transação inválida")
	ErrInvalidTransactionType      = errors.New("tipo de transação inválido")
	ErrItemDoesNotBelongToMonth    = errors.New("item não pertence a esta transação mensal")

	// Erros de negócio
	ErrCannotUpdateDeletedItem = errors.New("não é possível atualizar um item excluído")
	ErrCannotDeleteDeletedItem = errors.New("não é possível excluir um item já excluído")
	ErrNegativeAmount          = errors.New("valor não pode ser negativo")

	// Erros de Invoice
	ErrInvoiceProviderUnavailable = errors.New("provedor de fatura indisponível")
)
