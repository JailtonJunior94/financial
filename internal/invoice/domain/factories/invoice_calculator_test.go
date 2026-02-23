package factories_test

import (
	"testing"
	"time"

	"github.com/jailtonjunior94/financial/internal/invoice/domain/factories"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

// date é um helper para criar datas de forma legível nos testes.
func date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// refMonth é um helper para criar ReferenceMonth a partir de "YYYY-MM".
func refMonth(t *testing.T, value string) pkgVos.ReferenceMonth {
	t.Helper()
	rm, err := pkgVos.NewReferenceMonth(value)
	if err != nil {
		t.Fatalf("refMonth(%q): %v", value, err)
	}
	return rm
}

// TestCalculateInvoiceMonth valida a alocação determinística de ciclo de fatura.
//
// Regra formal:
//
//	C(v) = [ 25/(M-2), 24/(M-1) ]   (intervalo fechado)
//	offset = 1 se dia ≤ 24  →  vencimento = 01/(m+1)
//	offset = 2 se dia ≥ 25  →  vencimento = 01/(m+2)
func TestCalculateInvoiceMonth(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		purchaseDate time.Time
		wantDueDate  string // formato "YYYY-MM"
		justificativa string
	}{
		// ── Casos de borda em dia ──────────────────────────────────────────────
		{
			name:         "dia 01 — abertura do ciclo",
			purchaseDate: date(2024, 1, 1),
			wantDueDate:  "2024-02",
			justificativa: "01/Jan ∈ [25/Dez, 24/Jan] → Fev",
		},
		{
			name:         "dia 24 — último dia do ciclo (fronteira inclusiva ≤24)",
			purchaseDate: date(2024, 1, 24),
			wantDueDate:  "2024-02",
			justificativa: "24/Jan é o último dia de [25/Dez, 24/Jan] → Fev",
		},
		{
			name:         "dia 25 — primeiro dia do próximo ciclo (fronteira inclusiva ≥25)",
			purchaseDate: date(2024, 1, 25),
			wantDueDate:  "2024-03",
			justificativa: "25/Jan é o primeiro dia de [25/Jan, 24/Fev] → Mar",
		},
		{
			name:         "dia 31 — último dia de mês com 31 dias",
			purchaseDate: date(2024, 1, 31),
			wantDueDate:  "2024-03",
			justificativa: "31/Jan ∈ [25/Jan, 24/Fev] → Mar",
		},

		// ── Fevereiro — mês de 28 dias (não-bissexto) ─────────────────────────
		{
			name:         "fev/não-bissexto dia 24 — fronteira ≤24",
			purchaseDate: date(2023, 2, 24),
			wantDueDate:  "2023-03",
			justificativa: "24/Fev/2023 ∈ [25/Jan, 24/Fev] → Mar",
		},
		{
			name:         "fev/não-bissexto dia 25 — fronteira ≥25",
			purchaseDate: date(2023, 2, 25),
			wantDueDate:  "2023-04",
			justificativa: "25/Fev/2023 ∈ [25/Fev, 24/Mar] → Abr",
		},
		{
			name:         "fev/não-bissexto dia 28 — último dia",
			purchaseDate: date(2023, 2, 28),
			wantDueDate:  "2023-04",
			justificativa: "28/Fev/2023 ∈ [25/Fev, 24/Mar] → Abr",
		},

		// ── Fevereiro — mês de 29 dias (bissexto) ─────────────────────────────
		{
			name:         "fev/bissexto dia 24 — fronteira ≤24",
			purchaseDate: date(2024, 2, 24),
			wantDueDate:  "2024-03",
			justificativa: "24/Fev/2024 ∈ [25/Jan, 24/Fev] → Mar",
		},
		{
			name:         "fev/bissexto dia 25 — fronteira ≥25",
			purchaseDate: date(2024, 2, 25),
			wantDueDate:  "2024-04",
			justificativa: "25/Fev/2024 ∈ [25/Fev, 24/Mar] → Abr",
		},
		{
			name:         "fev/bissexto dia 29 — último dia",
			purchaseDate: date(2024, 2, 29),
			wantDueDate:  "2024-04",
			justificativa: "29/Fev/2024 ∈ [25/Fev, 24/Mar] → Abr",
		},

		// ── Mês de 30 dias ────────────────────────────────────────────────────
		// (Caso crítico: "7 dias antes" daria dia 25, não dia 24.)
		{
			name:         "abril dia 24 — mês com 30 dias, fronteira ≤24",
			purchaseDate: date(2024, 4, 24),
			wantDueDate:  "2024-05",
			justificativa: "24/Abr ∈ [25/Mar, 24/Abr] → Mai",
		},
		{
			name:         "abril dia 25 — mês com 30 dias, fronteira ≥25",
			purchaseDate: date(2024, 4, 25),
			wantDueDate:  "2024-06",
			justificativa: "25/Abr ∈ [25/Abr, 24/Mai] → Jun",
		},
		{
			name:         "setembro dia 30 — último dia do mês",
			purchaseDate: date(2024, 9, 30),
			wantDueDate:  "2024-11",
			justificativa: "30/Set ∈ [25/Set, 24/Out] → Nov",
		},

		// ── Virada de ano ─────────────────────────────────────────────────────
		{
			name:         "dezembro dia 24 — vencimento em janeiro do ano seguinte",
			purchaseDate: date(2024, 12, 24),
			wantDueDate:  "2025-01",
			justificativa: "24/Dez/2024 ∈ [25/Nov, 24/Dez] → Jan/2025",
		},
		{
			name:         "dezembro dia 25 — vencimento em fevereiro do ano seguinte",
			purchaseDate: date(2024, 12, 25),
			wantDueDate:  "2025-02",
			justificativa: "25/Dez/2024 ∈ [25/Dez/2024, 24/Jan/2025] → Fev/2025",
		},
		{
			name:         "dezembro dia 31 — réveillon, vencimento fevereiro",
			purchaseDate: date(2024, 12, 31),
			wantDueDate:  "2025-02",
			justificativa: "31/Dez/2024 ∈ [25/Dez/2024, 24/Jan/2025] → Fev/2025",
		},
		{
			name:         "novembro dia 30 — vencimento janeiro do ano seguinte",
			purchaseDate: date(2024, 11, 30),
			wantDueDate:  "2025-01",
			justificativa: "30/Nov/2024 ∈ [25/Nov, 24/Dez] → Jan/2025",
		},
		{
			name:         "outubro dia 31 — vencimento dezembro",
			purchaseDate: date(2024, 10, 31),
			wantDueDate:  "2024-12",
			justificativa: "31/Out ∈ [25/Out, 24/Nov] → Dez",
		},

		// ── Datas retroativas e futuras ────────────────────────────────────────
		{
			name:         "compra retroativa — 2022",
			purchaseDate: date(2022, 3, 15),
			wantDueDate:  "2022-04",
			justificativa: "15/Mar/2022 ≤ 24 → Abr/2022",
		},
		{
			name:         "compra futura — 2028",
			purchaseDate: date(2028, 11, 10),
			wantDueDate:  "2028-12",
			justificativa: "10/Nov/2028 ≤ 24 → Dez/2028",
		},

		// ── Ano 2000 (bissexto especial: divisível por 400) ───────────────────
		{
			name:         "ano 2000 fev/29 — bissexto secular",
			purchaseDate: date(2000, 2, 29),
			wantDueDate:  "2000-04",
			justificativa: "29/Fev/2000 ≥ 25 → Abr/2000",
		},

		// ── Casos extras da tabela exaustiva formal (Seção 5) ─────────────────
		{
			name:         "fev/bissexto dia 28 — ≥25 mas não último dia",
			purchaseDate: date(2024, 2, 28),
			wantDueDate:  "2024-04",
			justificativa: "28/Fev/2024 ≥ 25 ∈ [25/Fev, 24/Mar] → Abr",
		},
		{
			name:         "abril dia 30 — último dia de mês com 30 dias",
			purchaseDate: date(2024, 4, 30),
			wantDueDate:  "2024-06",
			justificativa: "30/Abr ∈ [25/Abr, 24/Mai] → Jun",
		},
		{
			name:         "dezembro dia 01 — primeiro dia do mês, virada de ano",
			purchaseDate: date(2024, 12, 1),
			wantDueDate:  "2025-01",
			justificativa: "01/Dez ≤ 24 ∈ [25/Nov, 24/Dez] → Jan/2025",
		},
		{
			name:         "novembro 2023 dia 25 — abertura cruzando virada de ano",
			purchaseDate: date(2023, 11, 25),
			wantDueDate:  "2024-01",
			justificativa: "25/Nov/2023 ≥ 25 ∈ [25/Nov, 24/Dez] → Jan/2024",
		},
		{
			name:         "dezembro 2023 dia 25 — abertura dupla virada de ano",
			purchaseDate: date(2023, 12, 25),
			wantDueDate:  "2024-02",
			justificativa: "25/Dez/2023 ≥ 25 ∈ [25/Dez/2023, 24/Jan/2024] → Fev/2024",
		},
		{
			name:         "novembro dia 01 — abertura do ciclo de dezembro",
			purchaseDate: date(2024, 11, 1),
			wantDueDate:  "2024-12",
			justificativa: "01/Nov ≤ 24 ∈ [25/Out, 24/Nov] → Dez",
		},
		{
			name:         "compra futura — maio 2026",
			purchaseDate: date(2026, 5, 10),
			wantDueDate:  "2026-06",
			justificativa: "10/Mai/2026 ≤ 24 → Jun/2026",
		},

		// ── Exemplos do enunciado original ────────────────────────────────────
		{
			name:         "enunciado: compra 21/01 → fatura 01/02",
			purchaseDate: date(2024, 1, 21),
			wantDueDate:  "2024-02",
			justificativa: "21/Jan ≤ 24 → Fev",
		},
		{
			name:         "enunciado: compra 27/01 → fatura 01/03",
			purchaseDate: date(2024, 1, 27),
			wantDueDate:  "2024-03",
			justificativa: "27/Jan ≥ 25 → Mar",
		},
	}

	calc := factories.NewInvoiceCalculator()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := calc.CalculateInvoiceMonth(tc.purchaseDate)
			want := refMonth(t, tc.wantDueDate)

			if !got.Equal(want) {
				t.Errorf("\ncompra:    %s\ngot:       %s\nwant:      %s\njustificativa: %s",
					tc.purchaseDate.Format("2006-01-02"),
					got.String(),
					want.String(),
					tc.justificativa,
				)
			}
		})
	}
}

// TestCalculateInvoiceDueDateIsFirstDay garante que D_venc = 1 é invariante.
func TestCalculateInvoiceDueDateIsFirstDay(t *testing.T) {
	t.Parallel()

	calc := factories.NewInvoiceCalculator()

	months := []string{"2024-01", "2024-02", "2024-12", "2025-01", "2023-02"}
	for _, m := range months {
		rm := refMonth(t, m)
		due := calc.CalculateDueDate(rm)
		if due.Day() != factories.DueDay {
			t.Errorf("CalculateDueDate(%s): got day=%d, want day=%d (D_venc)",
				m, due.Day(), factories.DueDay)
		}
		if due.Month() != rm.Month() || due.Year() != rm.Year() {
			t.Errorf("CalculateDueDate(%s): got %s, want same year/month", m, due.Format("2006-01-02"))
		}
	}
}

// TestCalculateInstallmentMonths valida parcelamento em múltiplos ciclos consecutivos.
func TestCalculateInstallmentMonths(t *testing.T) {
	t.Parallel()

	calc := factories.NewInvoiceCalculator()

	t.Run("à vista: 1 parcela", func(t *testing.T) {
		t.Parallel()
		months := calc.CalculateInstallmentMonths(date(2024, 1, 10), 1)
		if len(months) != 1 {
			t.Fatalf("got %d months, want 1", len(months))
		}
		want := refMonth(t, "2024-02")
		if !months[0].Equal(want) {
			t.Errorf("got %s, want %s", months[0], want)
		}
	})

	t.Run("3 parcelas: cruzam virada de ano", func(t *testing.T) {
		t.Parallel()
		// Compra 2024-11-10 (dia 10 ≤ 24) → primeira parcela Dez/2024
		// Parcelas: Dez/2024, Jan/2025, Fev/2025
		months := calc.CalculateInstallmentMonths(date(2024, 11, 10), 3)
		if len(months) != 3 {
			t.Fatalf("got %d months, want 3", len(months))
		}
		wants := []string{"2024-12", "2025-01", "2025-02"}
		for i, w := range wants {
			if !months[i].Equal(refMonth(t, w)) {
				t.Errorf("parcela[%d]: got %s, want %s", i+1, months[i], w)
			}
		}
	})

	t.Run("6 parcelas: compra no fechamento (dia 25)", func(t *testing.T) {
		t.Parallel()
		// Compra 2024-01-25 (dia 25 ≥ 25) → primeira parcela Mar/2024
		// Parcelas: Mar, Abr, Mai, Jun, Jul, Ago
		months := calc.CalculateInstallmentMonths(date(2024, 1, 25), 6)
		if len(months) != 6 {
			t.Fatalf("got %d months, want 6", len(months))
		}
		wants := []string{"2024-03", "2024-04", "2024-05", "2024-06", "2024-07", "2024-08"}
		for i, w := range wants {
			if !months[i].Equal(refMonth(t, w)) {
				t.Errorf("parcela[%d]: got %s, want %s", i+1, months[i], w)
			}
		}
	})

	t.Run("12 parcelas: cruzam ano completo", func(t *testing.T) {
		t.Parallel()
		// Compra 2024-12-25 → primeira parcela Fev/2025
		months := calc.CalculateInstallmentMonths(date(2024, 12, 25), 12)
		if len(months) != 12 {
			t.Fatalf("got %d months, want 12", len(months))
		}
		// Fev/2025 … Jan/2026
		wants := []string{
			"2025-02", "2025-03", "2025-04", "2025-05", "2025-06",
			"2025-07", "2025-08", "2025-09", "2025-10", "2025-11",
			"2025-12", "2026-01",
		}
		for i, w := range wants {
			if !months[i].Equal(refMonth(t, w)) {
				t.Errorf("parcela[%d]: got %s, want %s", i+1, months[i], w)
			}
		}
	})
}

// TestClosingDayConstant garante que a constante auditável está correta.
func TestClosingDayConstant(t *testing.T) {
	if factories.ClosingDay != 24 {
		t.Errorf("ClosingDay = %d, want 24 (D_fech contratual)", factories.ClosingDay)
	}
	if factories.DueDay != 1 {
		t.Errorf("DueDay = %d, want 1 (D_venc contratual)", factories.DueDay)
	}
}

// TestDueDateAlwaysAfterPurchaseDate verifica a invariante: vencimento > data de compra.
func TestDueDateAlwaysAfterPurchaseDate(t *testing.T) {
	t.Parallel()

	calc := factories.NewInvoiceCalculator()

	// Verifica todos os dias do ano 2024 (bissexto, cobertura máxima)
	start := date(2024, 1, 1)
	for d := 0; d < 366; d++ {
		purchaseDate := start.AddDate(0, 0, d)
		rm := calc.CalculateInvoiceMonth(purchaseDate)
		dueDate := calc.CalculateDueDate(rm)

		if !dueDate.After(purchaseDate) {
			t.Errorf("invariante violada: compra %s → vencimento %s (deve ser posterior)",
				purchaseDate.Format("2006-01-02"),
				dueDate.Format("2006-01-02"),
			)
		}
	}
}

// TestOpeningDayConstant garante que OpeningDay = ClosingDay + 1.
func TestOpeningDayConstant(t *testing.T) {
	if factories.OpeningDay != factories.ClosingDay+1 {
		t.Errorf("OpeningDay = %d, want ClosingDay+1 = %d",
			factories.OpeningDay, factories.ClosingDay+1)
	}
	if factories.OpeningDay != 25 {
		t.Errorf("OpeningDay = %d, want 25 (D_aber contratual)", factories.OpeningDay)
	}
}

// TestClosingDateFor valida que ClosingDateFor retorna o dia 24 do mês anterior
// ao vencimento, conforme definição formal: closing_date(v) = 24/month(abs_due-1).
func TestClosingDateFor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		referenceMonth string
		wantClosing    string
		justificativa  string
	}{
		{"2024-02", "2024-01-24", "Fev/2024 → fechamento 24/Jan/2024"},
		{"2024-03", "2024-02-24", "Mar/2024 → fechamento 24/Fev/2024 (fev. bissexto)"},
		{"2024-04", "2024-03-24", "Abr/2024 → fechamento 24/Mar/2024"},
		{"2024-05", "2024-04-24", "Mai/2024 → fechamento 24/Abr/2024 (mês de 30 dias)"},
		{"2024-12", "2024-11-24", "Dez/2024 → fechamento 24/Nov/2024"},
		{"2025-01", "2024-12-24", "Jan/2025 → fechamento 24/Dez/2024 (virada de ano)"},
		{"2025-02", "2025-01-24", "Fev/2025 → fechamento 24/Jan/2025"},
		{"2023-03", "2023-02-24", "Mar/2023 → fechamento 24/Fev/2023 (fev. não-bissexto)"},
	}

	calc := factories.NewInvoiceCalculator()

	for _, tc := range cases {
		t.Run(tc.referenceMonth, func(t *testing.T) {
			t.Parallel()

			rm := refMonth(t, tc.referenceMonth)
			got := calc.ClosingDateFor(rm)
			want, err := time.Parse("2006-01-02", tc.wantClosing)
			if err != nil {
				t.Fatalf("parse want date %q: %v", tc.wantClosing, err)
			}

			if !got.Equal(want) {
				t.Errorf("\nreferenceMonth: %s\ngot:  %s\nwant: %s\n%s",
					tc.referenceMonth,
					got.Format("2006-01-02"),
					want.Format("2006-01-02"),
					tc.justificativa,
				)
			}
			if got.Day() != factories.ClosingDay {
				t.Errorf("ClosingDateFor(%s).Day() = %d, want %d (ClosingDay)",
					tc.referenceMonth, got.Day(), factories.ClosingDay)
			}
		})
	}
}

// TestOpeningDateFor valida que OpeningDateFor retorna o dia 25 do mês
// anteanterior ao vencimento, conforme definição formal:
// opening_date(v) = 25/month(abs_due-2).
func TestOpeningDateFor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		referenceMonth string
		wantOpening    string
		justificativa  string
	}{
		{"2024-02", "2023-12-25", "Fev/2024 → abertura 25/Dez/2023"},
		{"2024-03", "2024-01-25", "Mar/2024 → abertura 25/Jan/2024"},
		{"2024-04", "2024-02-25", "Abr/2024 → abertura 25/Fev/2024 (bissexto, dia 25 existe)"},
		{"2024-05", "2024-03-25", "Mai/2024 → abertura 25/Mar/2024"},
		{"2024-12", "2024-10-25", "Dez/2024 → abertura 25/Out/2024"},
		{"2025-01", "2024-11-25", "Jan/2025 → abertura 25/Nov/2024 (virada de ano)"},
		{"2025-02", "2024-12-25", "Fev/2025 → abertura 25/Dez/2024"},
		{"2024-01", "2023-11-25", "Jan/2024 → abertura 25/Nov/2023"},
	}

	calc := factories.NewInvoiceCalculator()

	for _, tc := range cases {
		t.Run(tc.referenceMonth, func(t *testing.T) {
			t.Parallel()

			rm := refMonth(t, tc.referenceMonth)
			got := calc.OpeningDateFor(rm)
			want, err := time.Parse("2006-01-02", tc.wantOpening)
			if err != nil {
				t.Fatalf("parse want date %q: %v", tc.wantOpening, err)
			}

			if !got.Equal(want) {
				t.Errorf("\nreferenceMonth: %s\ngot:  %s\nwant: %s\n%s",
					tc.referenceMonth,
					got.Format("2006-01-02"),
					want.Format("2006-01-02"),
					tc.justificativa,
				)
			}
			if got.Day() != factories.OpeningDay {
				t.Errorf("OpeningDateFor(%s).Day() = %d, want %d (OpeningDay)",
					tc.referenceMonth, got.Day(), factories.OpeningDay)
			}
		})
	}
}

// TestNoDuplicateCycles verifica que dias consecutivos nas fronteiras do ciclo
// nunca produzem o mesmo vencimento.
func TestNoDuplicateCycles(t *testing.T) {
	t.Parallel()

	calc := factories.NewInvoiceCalculator()

	// Para cada fronteira dia 24→25, o vencimento deve avançar 1 mês.
	boundaries := []struct {
		before time.Time
		after  time.Time
	}{
		{date(2024, 1, 24), date(2024, 1, 25)},
		{date(2024, 2, 24), date(2024, 2, 25)},
		{date(2024, 3, 24), date(2024, 3, 25)},
		{date(2024, 4, 24), date(2024, 4, 25)},
		{date(2024, 11, 24), date(2024, 11, 25)},
		{date(2024, 12, 24), date(2024, 12, 25)},
	}

	for _, b := range boundaries {
		rmBefore := calc.CalculateInvoiceMonth(b.before)
		rmAfter := calc.CalculateInvoiceMonth(b.after)

		if rmBefore.Equal(rmAfter) {
			t.Errorf("fronteira %s/%s: ambos alocados em %s (devem ser faturas distintas)",
				b.before.Format("2006-01-02"),
				b.after.Format("2006-01-02"),
				rmBefore.String(),
			)
		}

		// O ciclo posterior deve ser exatamente 1 mês à frente.
		expectedAfter := rmBefore.AddMonths(1)
		if !rmAfter.Equal(expectedAfter) {
			t.Errorf("fronteira %s/%s: got after=%s, want %s (deve ser +1 mês)",
				b.before.Format("2006-01-02"),
				b.after.Format("2006-01-02"),
				rmAfter.String(),
				expectedAfter.String(),
			)
		}
	}
}

// TestCyclePartitionAllDays é um teste de propriedade que verifica, para todos
// os 366 dias do ano bissexto de 2024, que:
//
//  1. purchaseDate ∈ [openingDate, closingDate]  — sem lacunas
//  2. closingDate.Day() = ClosingDay (24)         — invariante de fechamento
//  3. openingDate.Day() = OpeningDay (25)         — invariante de abertura
//  4. openingDate ≤ closingDate                   — ordenação correta
//
// Prova computacional da Seção 4 (cobertura total e disjunção).
func TestCyclePartitionAllDays(t *testing.T) {
	t.Parallel()

	calc := factories.NewInvoiceCalculator()
	start := date(2024, 1, 1)

	for d := range 366 {
		p := start.AddDate(0, 0, d)
		opening, closing := calc.CycleFor(p)

		// Propriedade 1: p ∈ [opening, closing]
		if p.Before(opening) || p.After(closing) {
			t.Errorf("partição violada: %s ∉ [%s, %s]",
				p.Format("2006-01-02"),
				opening.Format("2006-01-02"),
				closing.Format("2006-01-02"),
			)
		}

		// Propriedade 2: fechamento é sempre o dia 24
		if closing.Day() != factories.ClosingDay {
			t.Errorf("closing day para %s: got %d, want %d",
				p.Format("2006-01-02"), closing.Day(), factories.ClosingDay)
		}

		// Propriedade 3: abertura é sempre o dia 25
		if opening.Day() != factories.OpeningDay {
			t.Errorf("opening day para %s: got %d, want %d",
				p.Format("2006-01-02"), opening.Day(), factories.OpeningDay)
		}

		// Propriedade 4: abertura ≤ fechamento
		if opening.After(closing) {
			t.Errorf("ordenação violada: opening %s > closing %s para compra %s",
				opening.Format("2006-01-02"),
				closing.Format("2006-01-02"),
				p.Format("2006-01-02"),
			)
		}
	}
}
