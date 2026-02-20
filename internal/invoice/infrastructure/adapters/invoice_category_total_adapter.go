package adapters

import (
	"context"
	"fmt"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	pkginterfaces "github.com/jailtonjunior94/financial/pkg/domain/interfaces"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

type invoiceCategoryTotalAdapter struct {
	invoiceRepository interfaces.InvoiceRepository
}

func NewInvoiceCategoryTotalAdapter(invoiceRepository interfaces.InvoiceRepository) pkginterfaces.InvoiceCategoryTotalProvider {
	return &invoiceCategoryTotalAdapter{invoiceRepository: invoiceRepository}
}

// GetCategoryTotal retorna a soma de InstallmentAmount das invoice_items de uma categoria no mês.
// A soma é feita em memória sobre as entidades carregadas, sem query SQL adicional.
func (a *invoiceCategoryTotalAdapter) GetCategoryTotal(
	ctx context.Context,
	userID sharedVos.UUID,
	referenceMonth pkgVos.ReferenceMonth,
	categoryID sharedVos.UUID,
) (sharedVos.Money, error) {
	invoices, err := a.invoiceRepository.FindByUserAndMonth(ctx, userID, referenceMonth)
	if err != nil {
		return sharedVos.Money{}, fmt.Errorf("failed to find invoices: %w", err)
	}

	total, _ := sharedVos.NewMoney(0, sharedVos.CurrencyBRL)

	for _, invoice := range invoices {
		for _, item := range invoice.Items {
			if item.CategoryID.String() != categoryID.String() {
				continue
			}
			sum, err := total.Add(item.InstallmentAmount)
			if err != nil {
				return sharedVos.Money{}, fmt.Errorf("failed to sum installment amounts: %w", err)
			}
			total = sum
		}
	}

	return total, nil
}
