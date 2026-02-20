package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/google/uuid"

	"github.com/jailtonjunior94/financial/internal/invoice/application/dtos"
	"github.com/jailtonjunior94/financial/pkg/money"
	"github.com/jailtonjunior94/financial/internal/invoice/domain"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/entities"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/events"
	"github.com/jailtonjunior94/financial/internal/invoice/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/outbox"
)

type (
	UpdatePurchaseUseCase interface {
		Execute(ctx context.Context, userID string, itemID string, input *dtos.PurchaseUpdateInput) (*dtos.PurchaseUpdateOutput, error)
	}

	updatePurchaseUseCase struct {
		uow           uow.UnitOfWork
		outboxService outbox.Service
		o11y          observability.Observability
	}
)

func NewUpdatePurchaseUseCase(
	uow uow.UnitOfWork,
	outboxService outbox.Service,
	o11y observability.Observability,
) UpdatePurchaseUseCase {
	return &updatePurchaseUseCase{
		uow:           uow,
		outboxService: outboxService,
		o11y:          o11y,
	}
}

func (u *updatePurchaseUseCase) Execute(ctx context.Context, userID string, itemID string, input *dtos.PurchaseUpdateInput) (*dtos.PurchaseUpdateOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "update_purchase_usecase.execute")
	defer span.End()

	// Parse itemID
	id, err := vos.NewUUIDFromString(itemID)
	if err != nil {
		return nil, fmt.Errorf("invalid item ID: %w", err)
	}

	// Parse categoryID
	categoryID, err := vos.NewUUIDFromString(input.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid category ID: %w", err)
	}

	// Collect updated items for response
	var updatedItems []*entities.InvoiceItem

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

		// Parse new total amount from string (half-even rounding)
		currency := items[0].TotalAmount.Currency()
		newTotalAmount, err := money.NewMoney(input.TotalAmount, currency)
		if err != nil {
			return fmt.Errorf("invalid total amount: %w", err)
		}

		newInstallmentAmount, err := newTotalAmount.Divide(int64(len(items)))
		if err != nil {
			return fmt.Errorf("failed to calculate installment amount: %w", err)
		}

		// Track affected invoices for total recalculation
		affectedInvoices := make(map[string]struct{})

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
			affectedInvoices[item.InvoiceID.String()] = struct{}{}
		}

		// Store updated items for response
		updatedItems = items

		// Recalculate totals for all affected invoices
		for invoiceIDStr := range affectedInvoices {
			invoiceID, _ := vos.NewUUIDFromString(invoiceIDStr)
			invoice, err := invoiceRepo.FindByID(ctx, invoiceID)
			if err != nil {
				return err
			}

			// Recalculate total from items
			var total vos.Money
			total, _ = vos.NewMoney(0, currency)

			for _, item := range invoice.Items {
				total, _ = total.Add(item.InstallmentAmount)
			}

			invoice.TotalAmount = total
			if err := invoiceRepo.Update(ctx, invoice); err != nil {
				return err
			}
		}

		// ✅ Salvar evento no outbox dentro da mesma transação
		// Collect affected months from the affected invoices
		monthsList := make([]string, 0, len(affectedInvoices))
		for invoiceIDStr := range affectedInvoices {
			invoiceID, _ := vos.NewUUIDFromString(invoiceIDStr)
			invoice, err := invoiceRepo.FindByID(ctx, invoiceID)
			if err == nil && invoice != nil {
				monthsList = append(monthsList, invoice.ReferenceMonth.String())
			}
		}

		// Converter userID string para UUID
		aggregateID, err := uuid.Parse(userID)
		if err != nil {
			return fmt.Errorf("invalid user_id: %w", err)
		}

		event := events.NewPurchaseUpdated(userID, input.CategoryID, monthsList)
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
			event.GetEventType(),
			payload,
		); err != nil {
			return fmt.Errorf("failed to save outbox event: %w", err)
		}

		return nil
	})

	if err != nil {
		u.o11y.Logger().Error(ctx, "failed to update purchase", observability.Error(err))
		return nil, err
	}

	u.o11y.Logger().Info(ctx, "purchase updated successfully",
		observability.String("user_id", userID),
		observability.Int("items", len(updatedItems)),
	)

	// Convert entities to DTOs
	itemOutputs := make([]dtos.InvoiceItemOutput, 0, len(updatedItems))
	for _, item := range updatedItems {
		installmentLabel := fmt.Sprintf("%d/%d", item.InstallmentNumber, item.InstallmentTotal)
		if item.InstallmentTotal == 1 {
			installmentLabel = "À vista"
		}

		itemOutputs = append(itemOutputs, dtos.InvoiceItemOutput{
			ID:                item.ID.String(),
			InvoiceID:         item.InvoiceID.String(),
			CategoryID:        item.CategoryID.String(),
			PurchaseDate:      item.PurchaseDate.Format("2006-01-02"),
			Description:       item.Description,
			TotalAmount:       fmt.Sprintf("%.2f", item.TotalAmount.Float()),
			InstallmentNumber: item.InstallmentNumber,
			InstallmentTotal:  item.InstallmentTotal,
			InstallmentAmount: fmt.Sprintf("%.2f", item.InstallmentAmount.Float()),
			InstallmentLabel:  installmentLabel,
			CreatedAt:         item.CreatedAt,
			UpdatedAt:         item.UpdatedAt.ValueOr(time.Time{}),
		})
	}

	return &dtos.PurchaseUpdateOutput{
		Items: itemOutputs,
	}, nil
}
