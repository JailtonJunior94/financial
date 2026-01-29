package usecase

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jailtonjunior94/financial/internal/invoice/application/dtos"
	"github.com/jailtonjunior94/financial/internal/invoice/domain"
	"github.com/jailtonjunior94/financial/internal/invoice/infrastructure/repositories"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	UpdatePurchaseUseCase interface {
		Execute(ctx context.Context, itemID string, input *dtos.PurchaseUpdateInput) error
	}

	updatePurchaseUseCase struct {
		uow  uow.UnitOfWork
		o11y observability.Observability
	}
)

func NewUpdatePurchaseUseCase(
	uow uow.UnitOfWork,
	o11y observability.Observability,
) UpdatePurchaseUseCase {
	return &updatePurchaseUseCase{
		uow:  uow,
		o11y: o11y,
	}
}

func (u *updatePurchaseUseCase) Execute(ctx context.Context, itemID string, input *dtos.PurchaseUpdateInput) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "update_purchase_usecase.execute")
	defer span.End()

	// Parse itemID
	id, err := vos.NewUUIDFromString(itemID)
	if err != nil {
		return fmt.Errorf("invalid item ID: %w", err)
	}

	// Parse categoryID
	categoryID, err := vos.NewUUIDFromString(input.CategoryID)
	if err != nil {
		return fmt.Errorf("invalid category ID: %w", err)
	}

	// Parse totalAmount
	totalAmountFloat, err := strconv.ParseFloat(input.TotalAmount, 64)
	if err != nil {
		return fmt.Errorf("invalid total amount format: %w", err)
	}

	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		invoiceRepo := repositories.NewInvoiceRepository(tx, u.o11y)

		// Find the item to get purchase origin details
		item, err := invoiceRepo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		if item == nil {
			return domain.ErrInvoiceItemNotFound
		}

		// Get the first item to extract purchase origin
		if len(item.Items) == 0 {
			return fmt.Errorf("no items found")
		}
		firstItem := item.Items[0]

		// Find all items from the same purchase
		items, err := invoiceRepo.FindItemsByPurchaseOrigin(
			ctx,
			firstItem.PurchaseDate.Format("2006-01-02"),
			firstItem.CategoryID,
			firstItem.Description,
		)
		if err != nil {
			return err
		}

		if len(items) == 0 {
			return fmt.Errorf("no items found for purchase")
		}

		// Parse new total amount and calculate new installment amount
		currency := items[0].TotalAmount.Currency()
		newTotalAmount, err := vos.NewMoneyFromFloat(totalAmountFloat, currency)
		if err != nil {
			return fmt.Errorf("invalid money value: %w", err)
		}

		newInstallmentAmount, err := newTotalAmount.Divide(int64(len(items)))
		if err != nil {
			return fmt.Errorf("failed to calculate installment amount: %w", err)
		}

		// Track affected invoices for total recalculation
		affectedInvoices := make(map[string]bool)

		// Update each item
		for _, item := range items {
			item.CategoryID = categoryID
			item.Description = input.Description
			item.TotalAmount = newTotalAmount
			item.InstallmentAmount = newInstallmentAmount

			if err := invoiceRepo.UpdateItem(ctx, item); err != nil {
				return err
			}

			// Track affected invoice
			affectedInvoices[item.InvoiceID.String()] = true
		}

		// Recalculate totals for all affected invoices
		for invoiceIDStr := range affectedInvoices {
			invoiceID, _ := vos.NewUUIDFromString(invoiceIDStr)
			invoice, err := invoiceRepo.FindByID(ctx, invoiceID)
			if err != nil {
				return err
			}

			// Recalculate total from items
			var total vos.Money
			total, _ = vos.NewMoneyFromFloat(0, currency)

			for _, item := range invoice.Items {
				total, _ = total.Add(item.InstallmentAmount)
			}

			invoice.TotalAmount = total
			if err := invoiceRepo.Update(ctx, invoice); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		u.o11y.Logger().Error(ctx, "failed to update purchase", observability.Error(err))
		return err
	}

	return nil
}
