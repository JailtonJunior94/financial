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

		// Agrupa itens por fatura (compras parceladas afetam N faturas)
		itemsByInvoice := make(map[string][]*entities.InvoiceItem)
		for _, item := range items {
			itemsByInvoice[item.InvoiceID.String()] = append(itemsByInvoice[item.InvoiceID.String()], item)
		}

		monthsList := make([]string, 0, len(itemsByInvoice))

		// Para cada fatura afetada: carrega o aggregate, remove via aggregate, persiste
		for invoiceIDStr, invoiceItems := range itemsByInvoice {
			invoiceID, _ := vos.NewUUIDFromString(invoiceIDStr)
			inv, err := invoiceRepo.FindByID(ctx, invoiceID)
			if err != nil {
				return err
			}
			if inv == nil {
				return domain.ErrInvoiceNotFound
			}

			for _, item := range invoiceItems {
				// Remoção via aggregate root — mantém invariantes e recalcula total
				if err := inv.RemoveItem(item.ID); err != nil {
					return err
				}

				// Persiste a remoção no banco
				if err := invoiceRepo.DeleteItem(ctx, item.ID); err != nil {
					return err
				}
			}

			// Persiste o total recalculado da fatura
			if err := invoiceRepo.Update(ctx, inv); err != nil {
				return err
			}

			monthsList = append(monthsList, inv.ReferenceMonth.String())
		}

		affectedMonths = monthsList

		// ✅ Salvar evento no outbox dentro da mesma transação
		aggregateID, err := uuid.Parse(userID)
		if err != nil {
			return fmt.Errorf("invalid user_id: %w", err)
		}

		event := events.NewPurchaseDeleted(userID, categoryID, affectedMonths)
		eventPayload := event.GetPayload().(events.PurchaseEventPayload)

		payload := outbox.JSONBPayload{
			"version":         eventPayload.Version,
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
