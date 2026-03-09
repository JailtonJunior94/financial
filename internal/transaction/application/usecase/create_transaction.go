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

	invoiceFactories "github.com/jailtonjunior94/financial/internal/invoice/domain/factories"
	invoiceInterfaces "github.com/jailtonjunior94/financial/internal/invoice/domain/interfaces"
	"github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/events"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/factories"
	transactionInterfaces "github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/outbox"
)

type (
	CreateTransactionUseCase interface {
		Execute(ctx context.Context, userID string, input *dtos.TransactionInput) ([]*dtos.TransactionOutput, error)
	}

	createTransactionUseCase struct {
		o11y            observability.Observability
		uow             uow.UnitOfWork
		repository      transactionInterfaces.TransactionRepository
		invoiceProvider transactionInterfaces.InvoiceProvider
		cardProvider    invoiceInterfaces.CardProvider
		factory         *factories.TransactionFactory
		outboxService   outbox.Service
	}
)

// NewCreateTransactionUseCase creates a new CreateTransactionUseCase.
func NewCreateTransactionUseCase(
	o11y observability.Observability,
	unitOfWork uow.UnitOfWork,
	repository transactionInterfaces.TransactionRepository,
	invoiceProvider transactionInterfaces.InvoiceProvider,
	cardProvider invoiceInterfaces.CardProvider,
	outboxService outbox.Service,
) CreateTransactionUseCase {
	return &createTransactionUseCase{
		o11y:            o11y,
		uow:             unitOfWork,
		repository:      repository,
		invoiceProvider: invoiceProvider,
		cardProvider:    cardProvider,
		factory:         factories.NewTransactionFactory(),
		outboxService:   outboxService,
	}
}

func (u *createTransactionUseCase) Execute(ctx context.Context, userID string, input *dtos.TransactionInput) ([]*dtos.TransactionOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "create_transaction_usecase.execute")
	defer span.End()

	if err := input.Validate(); err != nil {
		span.RecordError(err)
		return nil, err
	}

	userUUID, err := vos.NewUUIDFromString(userID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	transactionDate, err := time.Parse("2006-01-02", input.TransactionDate)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("invalid transaction_date: %w", err)
	}

	pm, _ := transactionVos.NewPaymentMethod(input.PaymentMethod)

	installments := input.Installments
	if installments <= 0 {
		installments = 1
	}

	var transactions []*entities.Transaction

	if pm.IsCredit() {
		cardUUID, err := vos.NewUUIDFromString(input.CardID)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("invalid card_id: %w", err)
		}

		billingInfo, err := u.cardProvider.GetCardBillingInfo(ctx, userUUID, cardUUID)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}

		calculator, err := invoiceFactories.NewInvoiceCalculator(billingInfo.DueDay, billingInfo.ClosingOffsetDays)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("invalid card billing configuration: %w", err)
		}

		months := calculator.CalculateInstallmentMonths(transactionDate, installments)
		invoiceIDs := make([]string, 0, installments)
		for _, month := range months {
			dueDate := calculator.CalculateDueDate(month)
			info, err := u.invoiceProvider.FindOrCreate(ctx, userUUID, cardUUID, month, dueDate)
			if err != nil {
				span.RecordError(err)
				return nil, err
			}
			invoiceIDs = append(invoiceIDs, info.ID.String())
		}

		if installments == 1 {
			createParams := factories.CreateParams{
				UserID:          userID,
				CategoryID:      input.CategoryID,
				SubcategoryID:   input.SubcategoryID,
				CardID:          input.CardID,
				InvoiceID:       invoiceIDs[0],
				Description:     input.Description,
				Amount:          input.Amount,
				PaymentMethod:   input.PaymentMethod,
				TransactionDate: transactionDate,
				Installments:    1,
			}
			tx, err := u.factory.Create(createParams)
			if err != nil {
				span.RecordError(err)
				return nil, err
			}
			transactions = []*entities.Transaction{tx}
		} else {
			installParams := factories.InstallmentParams{
				CreateParams: factories.CreateParams{
					UserID:          userID,
					CategoryID:      input.CategoryID,
					SubcategoryID:   input.SubcategoryID,
					CardID:          input.CardID,
					Description:     input.Description,
					Amount:          input.Amount,
					PaymentMethod:   input.PaymentMethod,
					TransactionDate: transactionDate,
					Installments:    installments,
				},
				InvoiceIDs: invoiceIDs,
			}
			transactions, err = u.factory.CreateInstallments(installParams)
			if err != nil {
				span.RecordError(err)
				return nil, err
			}
		}
	} else {
		createParams := factories.CreateParams{
			UserID:          userID,
			CategoryID:      input.CategoryID,
			SubcategoryID:   input.SubcategoryID,
			Description:     input.Description,
			Amount:          input.Amount,
			PaymentMethod:   input.PaymentMethod,
			TransactionDate: transactionDate,
			Installments:    1,
		}
		tx, err := u.factory.Create(createParams)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		transactions = []*entities.Transaction{tx}
	}

	err = u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		if err := u.repository.SaveAll(ctx, tx, transactions); err != nil {
			return err
		}
		for _, t := range transactions {
			referenceMonth := resolveReferenceMonth(t, transactionDate)
			event := events.NewTransactionCreatedEvent(
				t.ID,
				t.UserID,
				t.CategoryID,
				t.Amount,
				t.PaymentMethod,
				t.TransactionDate,
				referenceMonth,
				t.InvoiceID,
				t.InstallmentNumber,
				t.InstallmentTotal,
				t.InstallmentGroupID,
			)
			aggregateID, _ := uuid.Parse(t.ID.String())
			if err := u.outboxService.SaveDomainEvent(
				ctx,
				tx,
				aggregateID,
				"transaction",
				event.EventType(),
				outbox.JSONBPayload(event.Payload()),
			); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	u.o11y.Logger().Info(ctx, "request_completed",
		observability.String("operation", "CreateTransaction"),
		observability.String("layer", "usecase"),
		observability.String("entity", "transaction"),
		observability.String("user_id", userID),
	)

	return toOutputList(transactions), nil
}
