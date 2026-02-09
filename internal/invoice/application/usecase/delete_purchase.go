package usecase

import (
	"context"
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/google/uuid"

	"github.com/jailtonjunior94/financial/internal/invoice/domain"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/entities"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/events"
	"github.com/jailtonjunior94/financial/internal/invoice/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/outbox"
)

type (
	DeletePurchaseUseCase interface {
		Execute(ctx context.Context, userID string, itemID string, categoryID string) error
	}

	deletePurchaseUseCase struct {
		uow           uow.UnitOfWork
		outboxService outbox.Service
		o11y          observability.Observability
	}
)

func NewDeletePurchaseUseCase(
	uow uow.UnitOfWork,
	outboxService outbox.Service,
	o11y observability.Observability,
) DeletePurchaseUseCase {
	return &deletePurchaseUseCase{
		uow:           uow,
		outboxService: outboxService,
		o11y:          o11y,
	}
}

func (u *deletePurchaseUseCase) Execute(ctx context.Context, userID string, itemID string, categoryID string) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "delete_purchase_usecase.execute")
	defer span.End()

	// Parse itemID
	id, err := vos.NewUUIDFromString(itemID)
	if err != nil {
		return fmt.Errorf("invalid item ID: %w", err)
	}

	// Track affected months for sync
	var affectedMonths []string

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
		affectedInvoices := make(map[string]struct{})

		// Delete all items from the purchase
		for _, item := range items {
			if err := invoiceRepo.DeleteItem(ctx, item.ID); err != nil {
				return err
			}

			// Track affected invoice
			affectedInvoices[item.InvoiceID.String()] = struct{}{}
		}

		// Recalculate totals for all affected invoices
		currency := items[0].TotalAmount.Currency()
		affectedMonthsMap := make(map[string]bool)

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

			// Track affected month
			month := invoice.ReferenceMonth.String()
			affectedMonthsMap[month] = true
		}

		// Convert map to slice
		for month := range affectedMonthsMap {
			affectedMonths = append(affectedMonths, month)
		}

		// ✅ Salvar evento no outbox dentro da mesma transação
		aggregateID, err := uuid.Parse(userID)
		if err != nil {
			return fmt.Errorf("invalid user_id: %w", err)
		}

		event := events.NewPurchaseDeleted(userID, categoryID, affectedMonths)
		eventPayload := event.GetPayload().(events.PurchaseEventPayload)

		payload := outbox.JSONBPayload{
			"user_id":         eventPayload.UserID,
			"category_id":     eventPayload.CategoryID,
			"affected_months": eventPayload.AffectedMonths,
			"occurred_at":     eventPayload.OccurredAt,
		}

		if err := u.outboxService.SaveDomainEvent(
			ctx,
			tx,
			aggregateID,
			"invoice",
			events.PurchaseDeletedEventName,
			payload,
		); err != nil {
			return fmt.Errorf("failed to save outbox event: %w", err)
		}

		return nil
	})

	if err != nil {
		u.o11y.Logger().Error(ctx, "failed to delete purchase", observability.Error(err))
		return err
	}

	u.o11y.Logger().Info(ctx, "purchase deleted successfully",
		observability.String("user_id", userID),
		observability.Int("affected_months", len(affectedMonths)),
	)

	return nil
}
