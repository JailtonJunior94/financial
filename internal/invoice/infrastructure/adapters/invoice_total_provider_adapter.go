package adapters

import (
	"context"
	"fmt"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	pkginterfaces "github.com/jailtonjunior94/financial/pkg/domain/interfaces"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

// InvoiceTotalProviderAdapter implementa a interface InvoiceTotalProvider compartilhada.
// Este adapter segue o padrão Port & Adapter para integração entre módulos.
type InvoiceTotalProviderAdapter struct {
	invoiceRepository interfaces.InvoiceRepository
}

// NewInvoiceTotalProviderAdapter cria um novo adapter.
func NewInvoiceTotalProviderAdapter(invoiceRepository interfaces.InvoiceRepository) pkginterfaces.InvoiceTotalProvider {
	return &InvoiceTotalProviderAdapter{
		invoiceRepository: invoiceRepository,
	}
}

// GetClosedInvoiceTotal retorna o total das faturas do usuário no mês especificado.
//
// NOTA: Atualmente não há campo de status (open/closed) na entidade Invoice.
// Esta implementação soma TODAS as faturas do mês.
// Quando o campo Status for implementado, este método deverá filtrar apenas faturas fechadas.
func (a *InvoiceTotalProviderAdapter) GetClosedInvoiceTotal(
	ctx context.Context,
	userID sharedVos.UUID,
	referenceMonth pkgVos.ReferenceMonth,
) (sharedVos.Money, error) {
	// Buscar todas as faturas do usuário no mês
	// TODO: Implementar filtragem por status="closed"
	// Pendências para implementar este TODO:
	// 1. Adicionar migration: ALTER TABLE invoices ADD COLUMN status VARCHAR(20) DEFAULT 'open'
	// 2. Adicionar campo Status na entidade Invoice (domain/entities/invoice.go)
	// 3. Criar InvoiceStatus VO (domain/vos/invoice_status.go) com valores: open, closed, paid
	// 4. Adicionar parâmetro status no método FindByUserAndMonth do repository
	// 5. Atualizar esta chamada para: FindByUserAndMonth(ctx, userID, referenceMonth, InvoiceStatusClosed)
	invoices, err := a.invoiceRepository.FindByUserAndMonth(ctx, userID, referenceMonth)
	if err != nil {
		return sharedVos.Money{}, fmt.Errorf("failed to find invoices: %w", err)
	}

	// Se não houver faturas, retorna zero
	if len(invoices) == 0 {
		// Assumir BRL como moeda padrão (pode ser melhorado pegando da configuração do usuário)
		currency, _ := sharedVos.NewCurrency("BRL")
		zeroMoney, _ := sharedVos.NewMoney(0, currency)
		return zeroMoney, nil
	}

	// Somar os totais de todas as faturas
	// Usa a moeda da primeira fatura como referência
	total := invoices[0].TotalAmount
	for i := 1; i < len(invoices); i++ {
		sum, err := total.Add(invoices[i].TotalAmount)
		if err != nil {
			return sharedVos.Money{}, fmt.Errorf("failed to sum invoice totals: %w", err)
		}
		total = sum
	}

	return total, nil
}
