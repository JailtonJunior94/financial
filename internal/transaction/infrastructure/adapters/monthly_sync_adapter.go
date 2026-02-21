package adapters

import (
	"context"
	"fmt"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	invoiceInterfaces "github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	"github.com/jailtonjunior94/financial/internal/transaction/application/usecase"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

// MonthlySyncAdapter implementa a interface MonthlySyncService do módulo de invoice.
// Este adapter segue o padrão Port & Adapter para integração entre módulos.
type MonthlySyncAdapter struct {
	syncUseCase usecase.SyncMonthlyFromInvoicesUseCase
}

// NewMonthlySyncAdapter cria um novo adapter.
func NewMonthlySyncAdapter(syncUseCase usecase.SyncMonthlyFromInvoicesUseCase) invoiceInterfaces.MonthlySyncService {
	return &MonthlySyncAdapter{
		syncUseCase: syncUseCase,
	}
}

// SyncMonth sincroniza o MonthlyTransaction para o mês especificado.
func (a *MonthlySyncAdapter) SyncMonth(
	ctx context.Context,
	userID string,
	referenceMonth string,
	categoryID string,
) error {
	// Parse userID
	user, err := sharedVos.NewUUIDFromString(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Parse referenceMonth
	refMonth, err := pkgVos.NewReferenceMonth(referenceMonth)
	if err != nil {
		return fmt.Errorf("invalid reference month: %w", err)
	}

	// Parse categoryID
	catID, err := sharedVos.NewUUIDFromString(categoryID)
	if err != nil {
		return fmt.Errorf("invalid category ID: %w", err)
	}

	// Execute sync
	return a.syncUseCase.Execute(ctx, user, refMonth, catID)
}
