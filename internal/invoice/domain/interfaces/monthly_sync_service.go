package interfaces

import (
	"context"
)

// MonthlySyncService sincroniza MonthlyTransaction com base nas Invoices.
// Interface de domínio para integração com o módulo de transaction (Port & Adapter).
type MonthlySyncService interface {
	// SyncMonth atualiza o MonthlyTransaction para o mês/usuário especificado
	// baseado no total das invoices do mês.
	SyncMonth(ctx context.Context, userID string, referenceMonth string, categoryID string) error
}
