package adapters

import (
	"context"
	"fmt"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"

	budgetInterfaces "github.com/jailtonjunior94/financial/internal/budget/domain/interfaces"
	budgetVos "github.com/jailtonjunior94/financial/internal/budget/domain/vos"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	invoiceVos "github.com/jailtonjunior94/financial/internal/invoice/domain/vos"
)

type invoiceCategoryTotalAdapter struct {
	invoiceRepository interfaces.InvoiceRepository
}

func NewInvoiceCategoryTotalAdapter(invoiceRepository interfaces.InvoiceRepository) budgetInterfaces.InvoiceCategoryTotalProvider {
	return &invoiceCategoryTotalAdapter{invoiceRepository: invoiceRepository}
}

// GetCategoryTotal retorna a soma de InstallmentAmount das invoice_items de uma categoria no mês.
// A soma é feita em memória sobre as entidades carregadas, sem query SQL adicional.
func (a *invoiceCategoryTotalAdapter) GetCategoryTotal(
	ctx context.Context,
	userID sharedVos.UUID,
	referenceMonth budgetVos.ReferenceMonth,
	categoryID sharedVos.UUID,
) (sharedVos.Money, error) {
	invoiceRefMonth, err := invoiceVos.NewReferenceMonth(referenceMonth.String())
	if err != nil {
		return sharedVos.Money{}, fmt.Errorf("invalid reference month: %w", err)
	}

	invoices, err := a.invoiceRepository.FindByUserAndMonth(ctx, userID, invoiceRefMonth)
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
