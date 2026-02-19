package interfaces

import (
	"context"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	budgetVos "github.com/jailtonjunior94/financial/internal/budget/domain/vos"
)

// InvoiceCategoryTotalProvider fornece o total de parcelas de fatura por categoria e mês.
// Interface de domínio (Port) para integração com o módulo de invoice via Adapter.
type InvoiceCategoryTotalProvider interface {
	GetCategoryTotal(
		ctx context.Context,
		userID sharedVos.UUID,
		referenceMonth budgetVos.ReferenceMonth,
		categoryID sharedVos.UUID,
	) (sharedVos.Money, error)
}
