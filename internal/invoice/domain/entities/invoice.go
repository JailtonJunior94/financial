package entities

import (
	"slices"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/entity"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/invoice/domain"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

// Invoice é o Aggregate Root que representa uma fatura mensal de cartão.
type Invoice struct {
	entity.Base
	UserID         vos.UUID
	CardID         vos.UUID
	ReferenceMonth pkgVos.ReferenceMonth
	DueDate        time.Time
	TotalAmount    vos.Money
	Items          []*InvoiceItem
}

// NewInvoice cria uma nova fatura.
func NewInvoice(
	userID vos.UUID,
	cardID vos.UUID,
	referenceMonth pkgVos.ReferenceMonth,
	dueDate time.Time,
	currency vos.Currency,
) *Invoice {
	zeroMoney, _ := vos.NewMoney(0, currency)

	return &Invoice{
		UserID:         userID,
		CardID:         cardID,
		ReferenceMonth: referenceMonth,
		DueDate:        dueDate,
		TotalAmount:    zeroMoney,
		Items:          []*InvoiceItem{},
		Base: entity.Base{
			CreatedAt: time.Now().UTC(),
		},
	}
}

// AddItem adiciona um item à fatura e recalcula o total.
func (inv *Invoice) AddItem(item *InvoiceItem) error {
	if item == nil {
		return domain.ErrInvoiceItemNotFound
	}

	// Adiciona o item
	inv.Items = append(inv.Items, item)

	// Recalcula o total
	inv.recalculateTotalAmount()

	return nil
}

// AddItems adiciona múltiplos itens à fatura.
func (inv *Invoice) AddItems(items []*InvoiceItem) error {
	if len(items) == 0 {
		return domain.ErrInvoiceHasNoItems
	}

	inv.Items = append(inv.Items, items...)

	inv.recalculateTotalAmount()

	return nil
}

// RemoveItem remove um item da fatura e recalcula o total.
func (inv *Invoice) RemoveItem(itemID vos.UUID) error {
	originalLen := len(inv.Items)
	inv.Items = slices.DeleteFunc(inv.Items, func(item *InvoiceItem) bool {
		return item.ID.String() == itemID.String()
	})

	if len(inv.Items) == originalLen {
		return domain.ErrInvoiceItemNotFound
	}

	// Recalcula o total
	inv.recalculateTotalAmount()

	return nil
}

// FindItemByID busca um item pelo ID.
func (inv *Invoice) FindItemByID(itemID vos.UUID) *InvoiceItem {
	idx := slices.IndexFunc(inv.Items, func(item *InvoiceItem) bool {
		return item.ID.String() == itemID.String()
	})
	if idx == -1 {
		return nil
	}
	return inv.Items[idx]
}

// UpdateItemDetails atualiza os campos de um item via aggregate root e recalcula o total.
// Toda mutação de InvoiceItem deve passar por aqui para garantir consistência.
func (inv *Invoice) UpdateItemDetails(
	itemID vos.UUID,
	categoryID vos.UUID,
	description string,
	totalAmount vos.Money,
	installmentAmount vos.Money,
) error {
	item := inv.FindItemByID(itemID)
	if item == nil {
		return domain.ErrInvoiceItemNotFound
	}

	item.CategoryID = categoryID
	item.Description = description
	item.TotalAmount = totalAmount
	item.InstallmentAmount = installmentAmount
	item.UpdatedAt = vos.NewNullableTime(time.Now().UTC())

	inv.recalculateTotalAmount()
	inv.UpdatedAt = vos.NewNullableTime(time.Now().UTC())

	return nil
}

// RecalculateTotal recalcula o valor total da fatura com base nos itens.
func (inv *Invoice) RecalculateTotal() {
	inv.recalculateTotalAmount()
	inv.UpdatedAt = vos.NewNullableTime(time.Now().UTC())
}

// recalculateTotalAmount recalcula o valor total somando todas as parcelas.
func (inv *Invoice) recalculateTotalAmount() {
	zeroCurrency := inv.TotalAmount.Currency()
	total, err := vos.NewMoney(0, zeroCurrency)
	if err != nil {
		// Se falhar ao criar Money zero, mantem valor atual
		return
	}

	for _, item := range inv.Items {
		sum, err := total.Add(item.InstallmentAmount)
		if err != nil {
			// Se falhar ao somar, continua com próximo item
			continue
		}
		total = sum
	}

	inv.TotalAmount = total
}

// IsEmpty verifica se a fatura não tem itens.
func (inv *Invoice) IsEmpty() bool {
	return len(inv.Items) == 0
}

// ItemCount retorna a quantidade de itens na fatura.
func (inv *Invoice) ItemCount() int {
	return len(inv.Items)
}
