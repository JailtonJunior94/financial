package interfaces

import (
	"context"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

// InvoiceCategoryTotalProvider fornece o total de parcelas de fatura por categoria e mês.
// Interface compartilhada entre os módulos de invoice e budget (Port & Adapter).
type InvoiceCategoryTotalProvider interface {
	GetCategoryTotal(
		ctx context.Context,
		userID sharedVos.UUID,
		referenceMonth pkgVos.ReferenceMonth,
		categoryID sharedVos.UUID,
	) (sharedVos.Money, error)
}
