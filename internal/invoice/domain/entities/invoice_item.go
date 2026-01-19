package entities

import (
	"fmt"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/entity"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/invoice/domain"
)

// InvoiceItem representa uma compra/parcela lançada no cartão.
// Nota: Mutações devem passar pelo Invoice (aggregate root).
type InvoiceItem struct {
	entity.Base
	Invoice           *Invoice
	InvoiceID         vos.UUID
	CategoryID        vos.UUID
	PurchaseDate      time.Time
	Description       string
	TotalAmount       vos.Money // Valor total da compra original
	InstallmentNumber int       // Parcela atual (1 a N)
	InstallmentTotal  int       // Total de parcelas (1 para à vista)
	InstallmentAmount vos.Money // Valor desta parcela
}

// NewInvoiceItem cria um novo item de fatura com validações.
func NewInvoiceItem(
	invoice *Invoice,
	categoryID vos.UUID,
	purchaseDate time.Time,
	description string,
	totalAmount vos.Money,
	installmentNumber int,
	installmentTotal int,
	installmentAmount vos.Money,
) (*InvoiceItem, error) {
	// Validações de negócio
	if purchaseDate.After(time.Now()) {
		return nil, domain.ErrInvalidPurchaseDate
	}

	if description == "" {
		return nil, domain.ErrEmptyDescription
	}

	if totalAmount.IsNegative() || totalAmount.IsZero() {
		return nil, domain.ErrNegativeAmount
	}

	if installmentTotal < 1 {
		return nil, domain.ErrInvalidInstallmentTotal
	}

	if installmentNumber < 1 || installmentNumber > installmentTotal {
		return nil, domain.ErrInvalidInstallment
	}

	if installmentAmount.IsNegative() || installmentAmount.IsZero() {
		return nil, domain.ErrNegativeAmount
	}

	// Valida que o valor da parcela é correto (total / parcelas)
	// Para evitar problemas de arredondamento, validamos que a soma está correta
	expectedTotal, _ := installmentAmount.Multiply(int64(installmentTotal))
	if !expectedTotal.Equals(totalAmount) {
		// Permite uma diferença de centavos devido a arredondamento
		diff, _ := expectedTotal.Subtract(totalAmount)
		if diff.Abs().Cents() > int64(installmentTotal) {
			return nil, domain.ErrInstallmentAmountInvalid
		}
	}

	return &InvoiceItem{
		Invoice:           invoice,
		InvoiceID:         invoice.ID,
		CategoryID:        categoryID,
		PurchaseDate:      purchaseDate,
		Description:       description,
		TotalAmount:       totalAmount,
		InstallmentNumber: installmentNumber,
		InstallmentTotal:  installmentTotal,
		InstallmentAmount: installmentAmount,
		Base: entity.Base{
			CreatedAt: time.Now().UTC(),
		},
	}, nil
}

// IsInstallment retorna se este item é parcelado.
func (i *InvoiceItem) IsInstallment() bool {
	return i.InstallmentTotal > 1
}

// InstallmentLabel retorna a label da parcela (ex: "3/12").
func (i *InvoiceItem) InstallmentLabel() string {
	if !i.IsInstallment() {
		return "À vista"
	}
	return fmt.Sprintf("%d/%d", i.InstallmentNumber, i.InstallmentTotal)
}
