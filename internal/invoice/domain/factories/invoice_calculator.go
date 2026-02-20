package factories

import (
	"time"

	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

const (
	// ClosingDay é o dia fixo de fechamento da fatura — D_fech = 24.
	//
	// Contrato:
	//   dia_compra ≤ 24  →  pertence ao ciclo atual   (offset = 1)
	//   dia_compra ≥ 25  →  pertence ao ciclo seguinte (offset = 2)
	//
	// IMPORTANTE: "7 dias corridos antes do vencimento" é um mnemônico
	// válido apenas para meses com 30 dias. A regra canônica é esta constante.
	ClosingDay = 24

	// DueDay é o dia fixo de vencimento da fatura — D_venc = 1.
	DueDay = 1
)

// InvoiceCalculator implementa a regra determinística de alocação de ciclo de fatura.
//
// Modelo formal:
//
//	Ciclo C(v) para vencimento 01/M = [ 25/(M-2), 24/(M-1) ]  (intervalo fechado)
//
//	f(purchaseDate):
//	  offset = 1  se purchaseDate.Day() ≤ ClosingDay
//	  offset = 2  se purchaseDate.Day() ≥ 25
//	  invoice_due_date = 01 / (month + offset)   (aritmética ordinal de meses)
//
// Propriedades:
//   - Total:           toda LocalDate mapeia para exatamente 1 invoice_due_date
//   - Determinística:  sem dependência de clock, timezone ou estado externo
//   - Idempotente:     f(f(x)) ≠ f(x) por tipo, mas f é pura — mesmo input, mesmo output
//   - Auditável:       sem ramos ocultos, sem estado global
type InvoiceCalculator struct{}

// NewInvoiceCalculator cria um novo calculador de faturas.
func NewInvoiceCalculator() *InvoiceCalculator {
	return &InvoiceCalculator{}
}

// CalculateInvoiceMonth determina o mês de fatura (ReferenceMonth) de uma compra.
//
// A única decisão lógica é o comparador contra ClosingDay (24):
//
//	purchaseDate.Day() ≤ 24  →  invoice = 01/(m+1)
//	purchaseDate.Day() ≥ 25  →  invoice = 01/(m+2)
//
// Aritmética em ordinal de meses (base 0) garante correção em viradas de ano
// sem qualquer lógica especial para dezembro/janeiro.
func (c *InvoiceCalculator) CalculateInvoiceMonth(purchaseDate time.Time) pkgVos.ReferenceMonth {
	offset := 1
	if purchaseDate.Day() > ClosingDay {
		offset = 2
	}

	y := purchaseDate.Year()
	m := int(purchaseDate.Month())

	// Ordinal 0-indexed: elimina ambiguidade de virada de ano.
	// Invariante: totalMonths mod 12 ∈ [0, 11]
	totalMonths := y*12 + (m - 1) + offset
	dueYear := totalMonths / 12
	dueMonth := time.Month(totalMonths%12 + 1)

	return pkgVos.NewReferenceMonthFromDate(
		time.Date(dueYear, dueMonth, 1, 0, 0, 0, 0, time.UTC),
	)
}

// CalculateDueDate retorna a data de vencimento da fatura para um mês de referência.
//
// D_venc = 1: o vencimento é sempre o primeiro dia do mês de referência.
func (c *InvoiceCalculator) CalculateDueDate(referenceMonth pkgVos.ReferenceMonth) time.Time {
	return referenceMonth.FirstDay()
}

// CalculateInstallmentMonths retorna os ReferenceMonth de cada parcela de uma compra parcelada.
//
// A primeira parcela usa CalculateInvoiceMonth; as demais são meses consecutivos.
// O slice retornado tem exatamente installmentTotal elementos.
func (c *InvoiceCalculator) CalculateInstallmentMonths(
	purchaseDate time.Time,
	installmentTotal int,
) []pkgVos.ReferenceMonth {
	months := make([]pkgVos.ReferenceMonth, installmentTotal)

	firstMonth := c.CalculateInvoiceMonth(purchaseDate)
	months[0] = firstMonth

	for i := 1; i < installmentTotal; i++ {
		months[i] = firstMonth.AddMonths(i)
	}

	return months
}
