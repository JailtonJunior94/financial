package entities_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/card/domain/entities"
	cardVos "github.com/jailtonjunior94/financial/internal/card/domain/vos"
)

// CardBillingCycleTestSuite testa as regras brasileiras de faturamento.
type CardBillingCycleTestSuite struct {
	suite.Suite
}

func TestCardBillingCycleTestSuite(t *testing.T) {
	suite.Run(t, new(CardBillingCycleTestSuite))
}

// ===========================
// Testes de CalculateClosingDay
// ===========================

func (s *CardBillingCycleTestSuite) TestCalculateClosingDay_VencimentoDia10Offset7() {
	// Arrange: Vencimento dia 10, offset 7 → fechamento dia 3
	userID, _ := vos.NewUUID()
	name, _ := cardVos.NewCardName("Nubank")
	dueDay, _ := cardVos.NewDueDay(10)
	card, _ := entities.NewCard(userID, name, dueDay)
	offset, _ := cardVos.NewClosingOffsetDays(7)
	card.ClosingOffsetDays = offset

	// Act: Calcula fechamento para Janeiro/2025
	closingDate := card.CalculateClosingDay(2025, time.January)

	// Assert: Deve ser dia 3 de Janeiro
	s.Equal(2025, closingDate.Year())
	s.Equal(time.January, closingDate.Month())
	s.Equal(3, closingDate.Day())
}

func (s *CardBillingCycleTestSuite) TestCalculateClosingDay_VencimentoDia1Offset7_FechaNoMesAnterior() {
	// Arrange: Vencimento dia 1, offset 7 → fechamento dia 24 do mês anterior
	userID, _ := vos.NewUUID()
	name, _ := cardVos.NewCardName("BTG")
	dueDay, _ := cardVos.NewDueDay(1)
	card, _ := entities.NewCard(userID, name, dueDay)
	offset, _ := cardVos.NewClosingOffsetDays(7)
	card.ClosingOffsetDays = offset

	// Act: Calcula fechamento para Janeiro/2025 (vence dia 1/jan)
	closingDate := card.CalculateClosingDay(2025, time.January)

	// Assert: Deve ser dia 24 de Dezembro/2024
	s.Equal(2024, closingDate.Year())
	s.Equal(time.December, closingDate.Month())
	s.Equal(24, closingDate.Day())
}

func (s *CardBillingCycleTestSuite) TestCalculateClosingDay_VencimentoDia15Offset7() {
	// Arrange: Vencimento dia 15, offset 7 → fechamento dia 8
	userID, _ := vos.NewUUID()
	name, _ := cardVos.NewCardName("XP")
	dueDay, _ := cardVos.NewDueDay(15)
	card, _ := entities.NewCard(userID, name, dueDay)
	offset, _ := cardVos.NewClosingOffsetDays(7)
	card.ClosingOffsetDays = offset

	// Act
	closingDate := card.CalculateClosingDay(2025, time.March)

	// Assert: Deve ser dia 8 de Março
	s.Equal(2025, closingDate.Year())
	s.Equal(time.March, closingDate.Month())
	s.Equal(8, closingDate.Day())
}

// ===========================
// Testes de DetermineInvoiceMonth (CASOS REAIS BRASILEIROS)
// ===========================

func (s *CardBillingCycleTestSuite) TestDetermineInvoiceMonth_Nubank_Vencimento10_Offset7_CompraDia2() {
	// Arrange: Vencimento dia 10, offset 7 → fechamento dia 3
	userID, _ := vos.NewUUID()
	name, _ := cardVos.NewCardName("Nubank")
	dueDay, _ := cardVos.NewDueDay(10)
	card, _ := entities.NewCard(userID, name, dueDay)
	offset, _ := cardVos.NewClosingOffsetDays(7)
	card.ClosingOffsetDays = offset

	// Act: Compra no dia 02/jan (ANTES do fechamento dia 3)
	purchaseDate := time.Date(2025, time.January, 2, 12, 0, 0, 0, time.UTC)
	year, month := card.DetermineInvoiceMonth(purchaseDate)

	// Assert: Deve ir para fatura de Janeiro (vence 10/jan)
	s.Equal(2025, year)
	s.Equal(time.January, month)
}

func (s *CardBillingCycleTestSuite) TestDetermineInvoiceMonth_Nubank_Vencimento10_Offset7_CompraDia3() {
	// Arrange: Vencimento dia 10, offset 7 → fechamento dia 3
	userID, _ := vos.NewUUID()
	name, _ := cardVos.NewCardName("Nubank")
	dueDay, _ := cardVos.NewDueDay(10)
	card, _ := entities.NewCard(userID, name, dueDay)
	offset, _ := cardVos.NewClosingOffsetDays(7)
	card.ClosingOffsetDays = offset

	// Act: Compra NO DIA do fechamento (dia 3)
	purchaseDate := time.Date(2025, time.January, 3, 12, 0, 0, 0, time.UTC)
	year, month := card.DetermineInvoiceMonth(purchaseDate)

	// Assert: Deve ir para fatura de Fevereiro (vence 10/fev)
	s.Equal(2025, year)
	s.Equal(time.February, month)
}

func (s *CardBillingCycleTestSuite) TestDetermineInvoiceMonth_BTG_Vencimento1_Offset7_CompraDia23Dez() {
	// Arrange: Vencimento dia 1, offset 7 → fechamento dia 24 do mês anterior
	userID, _ := vos.NewUUID()
	name, _ := cardVos.NewCardName("BTG")
	dueDay, _ := cardVos.NewDueDay(1)
	card, _ := entities.NewCard(userID, name, dueDay)
	offset, _ := cardVos.NewClosingOffsetDays(7)
	card.ClosingOffsetDays = offset

	// Act: Compra dia 23/dez (ANTES do fechamento dia 24/dez)
	purchaseDate := time.Date(2024, time.December, 23, 12, 0, 0, 0, time.UTC)
	year, month := card.DetermineInvoiceMonth(purchaseDate)

	// Assert: Deve ir para fatura de Janeiro/2025 (vence 01/jan/2025)
	s.Equal(2025, year)
	s.Equal(time.January, month)
}

func (s *CardBillingCycleTestSuite) TestDetermineInvoiceMonth_BTG_Vencimento1_Offset7_CompraDia24Dez() {
	// Arrange: Vencimento dia 1, offset 7 → fechamento dia 24 do mês anterior
	userID, _ := vos.NewUUID()
	name, _ := cardVos.NewCardName("BTG")
	dueDay, _ := cardVos.NewDueDay(1)
	card, _ := entities.NewCard(userID, name, dueDay)
	offset, _ := cardVos.NewClosingOffsetDays(7)
	card.ClosingOffsetDays = offset

	// Act: Compra NO DIA do fechamento (dia 24/dez)
	purchaseDate := time.Date(2024, time.December, 24, 12, 0, 0, 0, time.UTC)
	year, month := card.DetermineInvoiceMonth(purchaseDate)

	// Assert: Deve ir para fatura de Fevereiro/2025 (vence 01/fev/2025)
	s.Equal(2025, year)
	s.Equal(time.February, month)
}

func (s *CardBillingCycleTestSuite) TestDetermineInvoiceMonth_XP_Vencimento15_Offset7_CompraDia7() {
	// Arrange: Vencimento dia 15, offset 7 → fechamento dia 8
	userID, _ := vos.NewUUID()
	name, _ := cardVos.NewCardName("XP")
	dueDay, _ := cardVos.NewDueDay(15)
	card, _ := entities.NewCard(userID, name, dueDay)
	offset, _ := cardVos.NewClosingOffsetDays(7)
	card.ClosingOffsetDays = offset

	// Act: Compra dia 7/mar (ANTES do fechamento dia 8)
	purchaseDate := time.Date(2025, time.March, 7, 12, 0, 0, 0, time.UTC)
	year, month := card.DetermineInvoiceMonth(purchaseDate)

	// Assert: Deve ir para fatura de Março (vence 15/mar)
	s.Equal(2025, year)
	s.Equal(time.March, month)
}

func (s *CardBillingCycleTestSuite) TestDetermineInvoiceMonth_XP_Vencimento15_Offset7_CompraDia8() {
	// Arrange: Vencimento dia 15, offset 7 → fechamento dia 8
	userID, _ := vos.NewUUID()
	name, _ := cardVos.NewCardName("XP")
	dueDay, _ := cardVos.NewDueDay(15)
	card, _ := entities.NewCard(userID, name, dueDay)
	offset, _ := cardVos.NewClosingOffsetDays(7)
	card.ClosingOffsetDays = offset

	// Act: Compra NO DIA do fechamento (dia 8)
	purchaseDate := time.Date(2025, time.March, 8, 12, 0, 0, 0, time.UTC)
	year, month := card.DetermineInvoiceMonth(purchaseDate)

	// Assert: Deve ir para fatura de Abril (vence 15/abr)
	s.Equal(2025, year)
	s.Equal(time.April, month)
}

// ===========================
// Teste de Determinismo: Usa < e NUNCA <=
// ===========================

func (s *CardBillingCycleTestSuite) TestDetermineInvoiceMonth_UsaOperadorMenorEstrict() {
	// Arrange: Testa que usa < e NUNCA <=
	userID, _ := vos.NewUUID()
	name, _ := cardVos.NewCardName("Test Card")
	dueDay, _ := cardVos.NewDueDay(20)
	card, _ := entities.NewCard(userID, name, dueDay)
	offset, _ := cardVos.NewClosingOffsetDays(5)
	card.ClosingOffsetDays = offset

	// Fechamento: dia 15 (20 - 5)

	// Act 1: Compra dia 14 (ANTES) → fatura atual
	purchase14 := time.Date(2025, time.January, 14, 23, 59, 59, 0, time.UTC)
	year1, month1 := card.DetermineInvoiceMonth(purchase14)

	// Act 2: Compra dia 15 (NO DIA) → próxima fatura
	purchase15 := time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC)
	year2, month2 := card.DetermineInvoiceMonth(purchase15)

	// Assert
	s.Equal(2025, year1)
	s.Equal(time.January, month1) // Fatura de janeiro

	s.Equal(2025, year2)
	s.Equal(time.February, month2) // Fatura de fevereiro
}

// ===========================
// Testes de Value Objects
// ===========================

func (s *CardBillingCycleTestSuite) TestClosingOffsetDays_ValorPadraoBrasileiro() {
	// Act
	offset := cardVos.NewDefaultClosingOffsetDays()

	// Assert: Padrão brasileiro é 7 dias
	s.Equal(7, offset.Int())
	s.True(offset.Valid)
}

func (s *CardBillingCycleTestSuite) TestClosingOffsetDays_ValidacaoMinimo() {
	// Act
	offset, err := cardVos.NewClosingOffsetDays(0)

	// Assert
	s.Error(err)
	s.False(offset.Valid)
	s.Contains(err.Error(), "invalid closing offset days")
}

func (s *CardBillingCycleTestSuite) TestClosingOffsetDays_ValidacaoMaximo() {
	// Act
	offset, err := cardVos.NewClosingOffsetDays(32)

	// Assert
	s.Error(err)
	s.False(offset.Valid)
}

func (s *CardBillingCycleTestSuite) TestClosingOffsetDays_ValorValido() {
	// Act
	offset, err := cardVos.NewClosingOffsetDays(10)

	// Assert
	s.NoError(err)
	s.Equal(10, offset.Int())
	s.True(offset.Valid)
}

// ===========================
// Teste de Integração: Ciclo Completo Nubank
// ===========================

func (s *CardBillingCycleTestSuite) TestCicloCompleto_CartaoNubank_MultipleCompras() {
	// Arrange: Simula um ciclo completo de faturamento Nubank
	// Vencimento dia 10, offset 7 → fechamento dia 3
	userID, _ := vos.NewUUID()
	name, _ := cardVos.NewCardName("Nubank")
	dueDay, _ := cardVos.NewDueDay(10)
	card, _ := entities.NewCard(userID, name, dueDay)
	offset, _ := cardVos.NewClosingOffsetDays(7)
	card.ClosingOffsetDays = offset

	// Compras no mês de Janeiro/2025
	compras := []struct {
		dia         int
		mesEsperado time.Month
		anoEsperado int
		descricao   string
	}{
		{1, time.January, 2025, "Compra dia 01 (antes fechamento)"},
		{2, time.January, 2025, "Compra dia 02 (antes fechamento)"},
		{3, time.February, 2025, "Compra dia 03 (NO DIA fechamento)"},
		{4, time.February, 2025, "Compra dia 04 (após fechamento)"},
		{9, time.February, 2025, "Compra dia 09 (antes próximo vencimento)"},
		{10, time.February, 2025, "Compra dia 10 (NO DIA vencimento)"},
		{11, time.February, 2025, "Compra dia 11 (após vencimento)"},
	}

	for _, compra := range compras {
		// Act
		purchaseDate := time.Date(2025, time.January, compra.dia, 12, 0, 0, 0, time.UTC)
		year, month := card.DetermineInvoiceMonth(purchaseDate)

		// Assert
		assert.Equal(s.T(), compra.anoEsperado, year, compra.descricao)
		assert.Equal(s.T(), compra.mesEsperado, month, compra.descricao)
	}
}

// ===========================
// Teste: Cartão com Offset Diferente do Padrão
// ===========================

func (s *CardBillingCycleTestSuite) TestDetermineInvoiceMonth_OffsetCustomizado() {
	// Arrange: Cartão com offset de 10 dias
	userID, _ := vos.NewUUID()
	name, _ := cardVos.NewCardName("Custom Card")
	dueDay, _ := cardVos.NewDueDay(25)
	card, _ := entities.NewCard(userID, name, dueDay)
	offset, _ := cardVos.NewClosingOffsetDays(10)
	card.ClosingOffsetDays = offset

	// Fechamento: dia 15 (25 - 10)

	// Act 1: Compra dia 14 → fatura atual
	purchase14 := time.Date(2025, time.January, 14, 12, 0, 0, 0, time.UTC)
	year1, month1 := card.DetermineInvoiceMonth(purchase14)

	// Act 2: Compra dia 15 → próxima fatura
	purchase15 := time.Date(2025, time.January, 15, 12, 0, 0, 0, time.UTC)
	year2, month2 := card.DetermineInvoiceMonth(purchase15)

	// Assert
	s.Equal(2025, year1)
	s.Equal(time.January, month1)

	s.Equal(2025, year2)
	s.Equal(time.February, month2)
}
