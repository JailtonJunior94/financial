package usecase

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/financial/internal/invoice/domain"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/entities"
	"github.com/jailtonjunior94/financial/internal/invoice/infrastructure/repositories"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	DeletePurchaseUseCase interface {
		Execute(ctx context.Context, itemID string) error
	}

	deletePurchaseUseCase struct {
		uow  uow.UnitOfWork
		o11y observability.Observability
	}
)

func NewDeletePurchaseUseCase(
	uow uow.UnitOfWork,
	o11y observability.Observability,
) DeletePurchaseUseCase {
	return &deletePurchaseUseCase{
		uow:  uow,
		o11y: o11y,
	}
}

func (u *deletePurchaseUseCase) Execute(ctx context.Context, itemID string) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "delete_purchase_usecase.execute")
	defer span.End()

	// Parse itemID
	id, err := vos.NewUUIDFromString(itemID)
	if err != nil {
		return fmt.Errorf("invalid item ID: %w", err)
	}

	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		invoiceRepo := repositories.NewInvoiceRepository(tx, u.o11y)

		// Find the invoice containing this item to get purchase origin details
		invoice, err := invoiceRepo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		if invoice == nil {
			return domain.ErrInvoiceNotFound
		}

		// Find the specific item
		var targetItem *entities.InvoiceItem
		for _, item := range invoice.Items {
			if item.ID.String() == itemID {
				targetItem = item
				break
			}
		}

		if targetItem == nil {
			return domain.ErrInvoiceItemNotFound
		}

		// Find all items from the same purchase origin
		items, err := invoiceRepo.FindItemsByPurchaseOrigin(
			ctx,
			targetItem.PurchaseDate.Format("2006-01-02"),
			targetItem.CategoryID,
			targetItem.Description,
		)
		if err != nil {
			return err
		}

		if len(items) == 0 {
			return fmt.Errorf("no items found for purchase")
		}

		// Track affected invoices for total recalculation
		affectedInvoices := make(map[string]bool)

		// Delete all items from the purchase
		for _, item := range items {
			if err := invoiceRepo.DeleteItem(ctx, item.ID); err != nil {
				return err
			}

			// Track affected invoice
			affectedInvoices[item.InvoiceID.String()] = true
		}

		// Recalculate totals for all affected invoices
		currency := items[0].TotalAmount.Currency()
		for invoiceIDStr := range affectedInvoices {
			invoiceID, _ := vos.NewUUIDFromString(invoiceIDStr)
			invoice, err := invoiceRepo.FindByID(ctx, invoiceID)
			if err != nil {
				return err
			}

			// Recalculate total from remaining items
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
		u.o11y.Logger().Error(ctx, "failed to delete purchase", observability.Error(err))
		return err
	}

	return nil
}
