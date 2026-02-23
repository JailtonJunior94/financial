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
	"github.com/jailtonjunior94/financial/internal/invoice/domain/entities"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/events"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/factories"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	"github.com/jailtonjunior94/financial/internal/invoice/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
	"github.com/jailtonjunior94/financial/pkg/outbox"
)

type (
	CreatePurchaseUseCase interface {
		Execute(ctx context.Context, userID string, input *dtos.PurchaseCreateInput) (*dtos.PurchaseCreateOutput, error)
	}

	createPurchaseUseCase struct {
		uow               uow.UnitOfWork
		cardProvider      interfaces.CardProvider
		outboxService     outbox.Service
		o11y              observability.Observability
		invoiceCalculator *factories.InvoiceCalculator
		fm                *metrics.FinancialMetrics
	}
)

func NewCreatePurchaseUseCase(
	uow uow.UnitOfWork,
	cardProvider interfaces.CardProvider,
	outboxService outbox.Service,
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
) CreatePurchaseUseCase {
	return &createPurchaseUseCase{
		uow:               uow,
		cardProvider:      cardProvider,
		outboxService:     outboxService,
		o11y:              o11y,
		invoiceCalculator: factories.NewInvoiceCalculator(),
		fm:                fm,
	}
}

func (u *createPurchaseUseCase) Execute(ctx context.Context, userID string, input *dtos.PurchaseCreateInput) (*dtos.PurchaseCreateOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "create_purchase_usecase.execute")
	defer span.End()

	// Parse userID
	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Parse cardID
	cardID, err := vos.NewUUIDFromString(input.CardID)
	if err != nil {
		return nil, fmt.Errorf("invalid card ID: %w", err)
	}

	// Parse categoryID
	categoryID, err := vos.NewUUIDFromString(input.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid category ID: %w", err)
	}

	// Parse purchaseDate
	purchaseDate, err := time.Parse("2006-01-02", input.PurchaseDate)
	if err != nil {
		return nil, fmt.Errorf("invalid purchase date format: %w", err)
	}

	// Parse totalAmount from string (half-even rounding)
	totalAmount, err := money.NewMoneyBRL(input.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid total amount: %w", err)
	}

	// ✅ Validar que o cartão pertence ao usuário via CardProvider (desacoplado)
	if _, err := u.cardProvider.GetCardBillingInfo(ctx, user, cardID); err != nil {
		u.o11y.Logger().Error(ctx, "failed to validate card ownership", observability.Error(err))
		return nil, err
	}

	// Calcular parcelas (valor de cada parcela)
	installmentAmount, err := totalAmount.Divide(int64(input.InstallmentTotal))
	if err != nil {
		return nil, fmt.Errorf("failed to calculate installment amount: %w", err)
	}

	// ✅ Usar InvoiceCalculator para determinar os meses de cada parcela
	installmentMonths := u.invoiceCalculator.CalculateInstallmentMonths(
		purchaseDate,
		input.InstallmentTotal,
	)

	// Collect created items for response
	createdItems := make([]*entities.InvoiceItem, 0, input.InstallmentTotal)

	// Criar os itens de fatura para cada parcela
	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		// Criar repositório com transação
		invoiceRepository := repositories.NewInvoiceRepository(tx, u.o11y, u.fm)

		// Para cada parcela, buscar ou criar a fatura de forma atômica
		for i := 0; i < input.InstallmentTotal; i++ {
			installmentNumber := i + 1
			referenceMonth := installmentMonths[i]

			// UpsertInvoice: INSERT ... ON CONFLICT DO UPDATE RETURNING
			// Atômico — elimina o race condition de criações concorrentes
			// para o mesmo (user_id, card_id, reference_month).
			dueDate := u.invoiceCalculator.CalculateDueDate(referenceMonth)
			newInvoice := entities.NewInvoice(user, cardID, referenceMonth, dueDate, vos.CurrencyBRL)
			invoiceID, _ := vos.NewUUID()
			newInvoice.SetID(invoiceID)

			invoice, err := invoiceRepository.UpsertInvoice(ctx, newInvoice)
			if err != nil {
				return fmt.Errorf("failed to upsert invoice: %w", err)
			}

			// Criar o item de fatura (parcela)
			item, err := entities.NewInvoiceItem(
				invoice.ID,
				categoryID,
				purchaseDate,
				input.Description,
				totalAmount,
				installmentNumber,
				input.InstallmentTotal,
				installmentAmount,
			)
			if err != nil {
				return fmt.Errorf("failed to create invoice item: %w", err)
			}

			// Gerar ID para o item
			itemID, _ := vos.NewUUID()
			item.SetID(itemID)

			// Adicionar item à fatura (recalcula total)
			if err := invoice.AddItem(item); err != nil {
				return err
			}

			// Persistir item
			if err := invoiceRepository.InsertItems(ctx, []*entities.InvoiceItem{item}); err != nil {
				return err
			}

			// Atualizar total da fatura
			if err := invoiceRepository.Update(ctx, invoice); err != nil {
				return err
			}

			// Collect created item
			createdItems = append(createdItems, item)
		}

		// ✅ Salvar evento no outbox dentro da mesma transação (ACID guarantee)
		monthsList := make([]string, 0, len(installmentMonths))
		for _, refMonth := range installmentMonths {
			monthsList = append(monthsList, refMonth.String())
		}

		// Converter userID para uuid.UUID
		aggregateID, err := uuid.Parse(userID)
		if err != nil {
			return fmt.Errorf("invalid user_id: %w", err)
		}

		// Criar evento e salvar no outbox
		event := events.NewPurchaseCreated(userID, input.CategoryID, monthsList)
		eventPayload := event.GetPayload().(events.PurchaseEventPayload)

		payload := outbox.JSONBPayload{
			"version":         eventPayload.Version,
			"user_id":         eventPayload.UserID,
			"category_id":     eventPayload.CategoryID,
			"affected_months": eventPayload.AffectedMonths,
			"occurred_at":     eventPayload.OccurredAt,
		}

		// Salvar no outbox - será processado assincronamente pelo worker
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
		u.o11y.Logger().Error(ctx, "failed to create purchase", observability.Error(err))
		return nil, err
	}

	// Event saved in outbox - will be processed by worker
	u.o11y.Logger().Info(ctx, "purchase created successfully",
		observability.String("user_id", userID),
		observability.Int("installments", len(createdItems)),
	)

	// Convert entities to DTOs
	itemOutputs := make([]dtos.InvoiceItemOutput, 0, len(createdItems))
	for _, item := range createdItems {
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

	return &dtos.PurchaseCreateOutput{
		Items: itemOutputs,
	}, nil
}

