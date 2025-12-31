package factories

import (
	"time"

	invoiceVos "github.com/jailtonjunior94/financial/internal/invoice/domain/vos"
)

// InvoiceCalculator contém a lógica de cálculo de fatura segundo padrão brasileiro
type InvoiceCalculator struct{}

// NewInvoiceCalculator cria um novo calculador de faturas
func NewInvoiceCalculator() *InvoiceCalculator {
	return &InvoiceCalculator{}
}

// CalculateInvoiceMonth calcula qual mês de fatura uma compra pertence
// baseado no dia de fechamento do cartão (padrão brasileiro)
//
// Regra:
// - Se purchaseDate.Day < closingDay → fatura do mês vigente
// - Se purchaseDate.Day >= closingDay → fatura do mês seguinte
//
// Exemplo:
// - Compra em 10/01, fechamento dia 25 → Fatura de Janeiro
// - Compra em 25/01, fechamento dia 25 → Fatura de Fevereiro
// - Compra em 26/01, fechamento dia 25 → Fatura de Fevereiro
func (c *InvoiceCalculator) CalculateInvoiceMonth(purchaseDate time.Time, closingDay int) invoiceVos.ReferenceMonth {
	year := purchaseDate.Year()
	month := purchaseDate.Month()

	// Se a compra foi feita no dia de fechamento ou depois,
	// ela entra na fatura do mês seguinte
	if purchaseDate.Day() >= closingDay {
		// Adiciona 1 mês
		nextMonth := purchaseDate.AddDate(0, 1, 0)
		year = nextMonth.Year()
		month = nextMonth.Month()
	}

	// Cria ReferenceMonth baseado no ano e mês calculados
	firstDayOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	return invoiceVos.NewReferenceMonthFromDate(firstDayOfMonth)
}

// CalculateDueDate calcula a data de vencimento da fatura
// baseado no mês de referência e no dia de vencimento do cartão
func (c *InvoiceCalculator) CalculateDueDate(referenceMonth invoiceVos.ReferenceMonth, dueDay int) time.Time {
	year := referenceMonth.Year()
	month := referenceMonth.Month()

	// Cria a data de vencimento no dia especificado do mês
	dueDate := time.Date(year, month, dueDay, 0, 0, 0, 0, time.UTC)

	// Se o dia não existe no mês (ex: 31 de fevereiro), Go ajusta automaticamente
	// para o último dia válido do mês
	return dueDate
}

// CalculateInstallmentMonths calcula os meses de referência para cada parcela
// de uma compra parcelada
//
// Retorna um slice com os ReferenceMonth de cada parcela
func (c *InvoiceCalculator) CalculateInstallmentMonths(
	purchaseDate time.Time,
	closingDay int,
	installmentTotal int,
) []invoiceVos.ReferenceMonth {
	months := make([]invoiceVos.ReferenceMonth, installmentTotal)

	// Primeira parcela: calcula normalmente
	firstMonth := c.CalculateInvoiceMonth(purchaseDate, closingDay)
	months[0] = firstMonth

	// Demais parcelas: meses subsequentes
	for i := 1; i < installmentTotal; i++ {
		months[i] = firstMonth.AddMonths(i)
	}

	return months
}
