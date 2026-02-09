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
	"github.com/jailtonjunior94/financial/internal/invoice/domain/entities"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/events"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/factories"
	"github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	invoiceVos "github.com/jailtonjunior94/financial/internal/invoice/domain/vos"
	"github.com/jailtonjunior94/financial/internal/invoice/infrastructure/repositories"
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
	}
)

func NewCreatePurchaseUseCase(
	uow uow.UnitOfWork,
	cardProvider interfaces.CardProvider,
	outboxService outbox.Service,
	o11y observability.Observability,
) CreatePurchaseUseCase {
	return &createPurchaseUseCase{
		uow:               uow,
		cardProvider:      cardProvider,
		outboxService:     outboxService,
		o11y:              o11y,
		invoiceCalculator: factories.NewInvoiceCalculator(),
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

	// Parse totalAmount from string (preserves precision)
	totalAmount, err := vos.NewMoneyFromString(input.TotalAmount, vos.CurrencyBRL)
	if err != nil {
		return nil, fmt.Errorf("invalid total amount: %w", err)
	}

	// ✅ Obter informações de faturamento do cartão via CardProvider (desacoplado)
	cardBillingInfo, err := u.cardProvider.GetCardBillingInfo(ctx, user, cardID)
	if err != nil {
		u.o11y.Logger().Error(ctx, "failed to get card billing info", observability.Error(err))
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
		cardBillingInfo.DueDay,
		cardBillingInfo.ClosingOffsetDays,
		input.InstallmentTotal,
	)

	// Collect created items for response
	createdItems := make([]*entities.InvoiceItem, 0, input.InstallmentTotal)

	// Criar os itens de fatura para cada parcela
	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		// Criar repositório com transação
		invoiceRepository := repositories.NewInvoiceRepository(tx, u.o11y)

		// Para cada parcela, criar ou buscar a fatura e adicionar o item
		for i := 0; i < input.InstallmentTotal; i++ {
			installmentNumber := i + 1
			referenceMonth := installmentMonths[i]

			// Buscar ou criar a fatura para este mês
			invoice, err := u.findOrCreateInvoice(
				ctx,
				invoiceRepository,
				user,
				cardID,
				referenceMonth,
				cardBillingInfo.DueDay,
				vos.CurrencyBRL,
			)
			if err != nil {
				return fmt.Errorf("failed to find or create invoice: %w", err)
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
			TotalAmount:       item.TotalAmount.String(),
			InstallmentNumber: item.InstallmentNumber,
			InstallmentTotal:  item.InstallmentTotal,
			InstallmentAmount: item.InstallmentAmount.String(),
			InstallmentLabel:  installmentLabel,
			CreatedAt:         item.CreatedAt,
			UpdatedAt:         item.UpdatedAt.ValueOr(time.Time{}),
		})
	}

	return &dtos.PurchaseCreateOutput{
		Items: itemOutputs,
	}, nil
}

// findOrCreateInvoice busca ou cria uma fatura para o mês de referência.
func (u *createPurchaseUseCase) findOrCreateInvoice(
	ctx context.Context,
	repo interfaces.InvoiceRepository,
	userID vos.UUID,
	cardID vos.UUID,
	referenceMonth invoiceVos.ReferenceMonth,
	dueDay int,
	currency vos.Currency,
) (*entities.Invoice, error) {
	// Tentar buscar fatura existente
	invoice, err := repo.FindByUserAndCardAndMonth(ctx, userID, cardID, referenceMonth)
	if err != nil {
		return nil, err
	}

	// Se encontrou, retorna
	if invoice != nil {
		return invoice, nil
	}

	// Se não encontrou, cria nova fatura
	dueDate := u.invoiceCalculator.CalculateDueDate(referenceMonth, dueDay)

	invoice = entities.NewInvoice(userID, cardID, referenceMonth, dueDate, currency)

	// Gerar ID
	invoiceID, _ := vos.NewUUID()
	invoice.SetID(invoiceID)

	// Persistir
	if err := repo.Insert(ctx, invoice); err != nil {
		return nil, err
	}

	return invoice, nil
}
