package factories

import (
	"time"

	invoiceVos "github.com/jailtonjunior94/financial/internal/invoice/domain/vos"
)

// InvoiceCalculator contém a lógica de cálculo de fatura segundo padrão brasileiro.
type InvoiceCalculator struct{}

// NewInvoiceCalculator cria um novo calculador de faturas.
func NewInvoiceCalculator() *InvoiceCalculator {
	return &InvoiceCalculator{}
}

// calculateClosingDay calcula o dia de fechamento da fatura.
// Usa a mesma lógica do módulo Card (fonte da verdade).
// Regra: closingDay = dueDay - closingOffsetDays.
// Se resultado <= 0, volta para o mês anterior.
func (c *InvoiceCalculator) calculateClosingDay(referenceYear int, referenceMonth time.Month, dueDay int, closingOffsetDays int) time.Time {
	closingDay := dueDay - closingOffsetDays

	// Se ficou negativo ou zero, volta para o mês anterior
	if closingDay <= 0 {
		// Pega o último dia do mês anterior
		firstDayOfReferenceMonth := time.Date(referenceYear, referenceMonth, 1, 0, 0, 0, 0, time.UTC)
		lastDayOfPreviousMonth := firstDayOfReferenceMonth.AddDate(0, 0, -1)

		// Calcula o dia de fechamento no mês anterior
		// Exemplo: vence dia 1, offset 7 → fecha dia 24 do mês anterior (31 - 7 = 24)
		closingDay = lastDayOfPreviousMonth.Day() - closingOffsetDays

		return time.Date(referenceYear, referenceMonth, 1, 0, 0, 0, 0, time.UTC).AddDate(0, -1, closingDay-1)
	}

	return time.Date(referenceYear, referenceMonth, closingDay, 0, 0, 0, 0, time.UTC)
}

// CalculateInvoiceMonth calcula qual mês de fatura uma compra pertence.
// baseado nas regras brasileiras de faturamento.
//
// Regra CRÍTICA:
// - Se purchaseDate < closingDate → fatura do mês vigente.
// - Se purchaseDate >= closingDate → fatura do mês seguinte.
//
// Usa < e NUNCA <= (determinístico).
func (c *InvoiceCalculator) CalculateInvoiceMonth(purchaseDate time.Time, dueDay int, closingOffsetDays int) invoiceVos.ReferenceMonth {
	// Calcula o vencimento potencial no mês da compra
	dueDate := time.Date(purchaseDate.Year(), purchaseDate.Month(), dueDay, 0, 0, 0, 0, time.UTC)

	// Se o vencimento do mês da compra já passou, olha para o próximo mês
	// Exemplo: compra dia 24/dez com vencimento dia 1 → próximo vencimento é 01/jan
	if dueDate.Before(purchaseDate) {
		dueDate = dueDate.AddDate(0, 1, 0)
	}

	// Calcula a data de fechamento para essa fatura
	closingDate := c.calculateClosingDay(dueDate.Year(), dueDate.Month(), dueDay, closingOffsetDays)

	// Regra determinística: usa < e NUNCA <=
	if purchaseDate.Before(closingDate) {
		// Compra ANTES do fechamento → vai para esta fatura
		firstDayOfMonth := time.Date(dueDate.Year(), dueDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		return invoiceVos.NewReferenceMonthFromDate(firstDayOfMonth)
	}

	// Compra NO DIA ou APÓS o fechamento → vai para a fatura do próximo mês
	nextDueDate := dueDate.AddDate(0, 1, 0)
	firstDayOfNextMonth := time.Date(nextDueDate.Year(), nextDueDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	return invoiceVos.NewReferenceMonthFromDate(firstDayOfNextMonth)
}

// CalculateDueDate calcula a data de vencimento da fatura.
// baseado no mês de referência e no dia de vencimento do cartão.
func (c *InvoiceCalculator) CalculateDueDate(referenceMonth invoiceVos.ReferenceMonth, dueDay int) time.Time {
	year := referenceMonth.Year()
	month := referenceMonth.Month()

	// Cria a data de vencimento no dia especificado do mês
	dueDate := time.Date(year, month, dueDay, 0, 0, 0, 0, time.UTC)

	// Se o dia não existe no mês (ex: 31 de fevereiro), Go ajusta automaticamente
	// para o último dia válido do mês
	return dueDate
}

// CalculateInstallmentMonths calcula os meses de referência para cada parcela.
// de uma compra parcelada.
//
// Retorna um slice com os ReferenceMonth de cada parcela.
func (c *InvoiceCalculator) CalculateInstallmentMonths(
	purchaseDate time.Time,
	dueDay int,
	closingOffsetDays int,
	installmentTotal int,
) []invoiceVos.ReferenceMonth {
	months := make([]invoiceVos.ReferenceMonth, installmentTotal)

	// Primeira parcela: calcula normalmente
	firstMonth := c.CalculateInvoiceMonth(purchaseDate, dueDay, closingOffsetDays)
	months[0] = firstMonth

	// Demais parcelas: meses subsequentes
	for i := 1; i < installmentTotal; i++ {
		months[i] = firstMonth.AddMonths(i)
	}

	return months
}
