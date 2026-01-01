package entities

import (
	"time"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/card/domain/vos"
)

type Card struct {
	ID                 sharedVos.UUID
	UserID             sharedVos.UUID
	Name               vos.CardName
	DueDay             vos.DueDay
	ClosingOffsetDays  vos.ClosingOffsetDays
	CreatedAt          sharedVos.NullableTime
	UpdatedAt          sharedVos.NullableTime
	DeletedAt          sharedVos.NullableTime
}

func NewCard(userID sharedVos.UUID, name vos.CardName, dueDay vos.DueDay) (*Card, error) {
	card := &Card{
		Name:              name,
		UserID:            userID,
		DueDay:            dueDay,
		ClosingOffsetDays: vos.NewDefaultClosingOffsetDays(), // Padrão brasileiro: 7 dias
		CreatedAt:         sharedVos.NewNullableTime(time.Now()),
	}
	return card, nil
}

func (c *Card) Update(name string, dueDay int, closingOffsetDays int) error {
	cardName, err := vos.NewCardName(name)
	if err != nil {
		return err
	}

	cardDueDay, err := vos.NewDueDay(dueDay)
	if err != nil {
		return err
	}

	offset, err := vos.NewClosingOffsetDays(closingOffsetDays)
	if err != nil {
		return err
	}

	c.Name = cardName
	c.DueDay = cardDueDay
	c.ClosingOffsetDays = offset
	c.UpdatedAt = sharedVos.NewNullableTime(time.Now())

	return nil
}

func (c *Card) Delete() *Card {
	c.DeletedAt = sharedVos.NewNullableTime(time.Now())
	return c
}

// CalculateClosingDay calcula o dia de fechamento da fatura para um dado mês.
// Regra brasileira: fechamento = vencimento - offset.
// Se resultado <= 0, volta para o mês anterior.
func (c *Card) CalculateClosingDay(referenceYear int, referenceMonth time.Month) time.Time {
	dueDay := c.DueDay.Int()
	offset := c.ClosingOffsetDays.Int()

	// Calcula dia de fechamento
	closingDay := dueDay - offset

	// Se ficou negativo ou zero, volta para o mês anterior
	if closingDay <= 0 {
		// Pega o último dia do mês anterior
		firstDayOfReferenceMonth := time.Date(referenceYear, referenceMonth, 1, 0, 0, 0, 0, time.UTC)
		lastDayOfPreviousMonth := firstDayOfReferenceMonth.AddDate(0, 0, -1)

		// Calcula o dia de fechamento no mês anterior
		// Exemplo: vence dia 1, offset 7 → fecha dia 24 do mês anterior (31 - 7 = 24)
		closingDay = lastDayOfPreviousMonth.Day() - offset

		return time.Date(referenceYear, referenceMonth, 1, 0, 0, 0, 0, time.UTC).AddDate(0, -1, closingDay-1)
	}

	return time.Date(referenceYear, referenceMonth, closingDay, 0, 0, 0, 0, time.UTC)
}

// DetermineInvoiceMonth determina a qual fatura uma compra pertence.
// Regra CRÍTICA (padrão brasileiro):
// - Se purchaseDate < closingDate → fatura do mês de vencimento.
// - Se purchaseDate >= closingDate → fatura do próximo mês.
//
// Exemplo 1: Vencimento dia 10, offset 7 (fechamento dia 3).
// - Compra dia 02/jan → fatura de janeiro (vence 10/jan).
// - Compra dia 03/jan → fatura de fevereiro (vence 10/fev).
//
// Exemplo 2: Vencimento dia 01, offset 7 (fechamento dia 24 do mês anterior).
// - Compra dia 23/dez → fatura de janeiro (vence 01/jan).
// - Compra dia 24/dez → fatura de fevereiro (vence 01/fev).
func (c *Card) DetermineInvoiceMonth(purchaseDate time.Time) (year int, month time.Month) {
	dueDay := c.DueDay.Int()

	// Calcula o vencimento potencial no mês da compra
	dueDate := time.Date(purchaseDate.Year(), purchaseDate.Month(), dueDay, 0, 0, 0, 0, time.UTC)

	// Se o vencimento do mês da compra já passou, olha para o próximo mês
	// Exemplo: compra dia 24/dez com vencimento dia 1 → próximo vencimento é 01/jan
	if dueDate.Before(purchaseDate) {
		dueDate = dueDate.AddDate(0, 1, 0)
	}

	// Calcula a data de fechamento para essa fatura
	closingDate := c.CalculateClosingDay(dueDate.Year(), dueDate.Month())

	// Regra determinística: usa < e NUNCA <=
	if purchaseDate.Before(closingDate) {
		// Compra ANTES do fechamento → vai para esta fatura
		return dueDate.Year(), dueDate.Month()
	}

	// Compra NO DIA ou APÓS o fechamento → vai para a fatura do próximo mês
	nextDueDate := dueDate.AddDate(0, 1, 0)
	return nextDueDate.Year(), nextDueDate.Month()
}
