package usecase

import (
	"context"
	"fmt"

	appstrategies "github.com/jailtonjunior94/financial/internal/transaction/application/strategies"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/uow"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	SyncMonthlyFromInvoicesUseCase interface {
		Execute(ctx context.Context, userID sharedVos.UUID, referenceMonth transactionVos.ReferenceMonth, categoryID sharedVos.UUID) error
	}

	syncMonthlyFromInvoicesUseCase struct {
		uow                  uow.UnitOfWork
		repo                 interfaces.TransactionRepository
		invoiceTotalProvider interfaces.InvoiceTotalProvider
		ccItemPersister      appstrategies.CreditCardItemPersister
		o11y                 observability.Observability
	}
)

func NewSyncMonthlyFromInvoicesUseCase(
	uow uow.UnitOfWork,
	repo interfaces.TransactionRepository,
	invoiceTotalProvider interfaces.InvoiceTotalProvider,
	ccItemPersister appstrategies.CreditCardItemPersister,
	o11y observability.Observability,
) SyncMonthlyFromInvoicesUseCase {
	return &syncMonthlyFromInvoicesUseCase{
		uow:                  uow,
		repo:                 repo,
		invoiceTotalProvider: invoiceTotalProvider,
		ccItemPersister:      ccItemPersister,
		o11y:                 o11y,
	}
}

func (u *syncMonthlyFromInvoicesUseCase) Execute(
	ctx context.Context,
	userID sharedVos.UUID,
	referenceMonth transactionVos.ReferenceMonth,
	categoryID sharedVos.UUID,
) error {
	ctx, span := u.o11y.Tracer().Start(ctx, "sync_monthly_from_invoices_usecase.execute")
	defer span.End()

	err := u.uow.Do(ctx, func(ctx context.Context, tx database.DBTX) error {
		invoiceTotal, err := u.invoiceTotalProvider.GetClosedInvoiceTotal(ctx, userID, referenceMonth)
		if err != nil {
			return fmt.Errorf("failed to get invoice total: %w", err)
		}

		u.o11y.Logger().Info(ctx, "fetched invoice total for month",
			observability.String("user_id", userID.String()),
			observability.String("reference_month", referenceMonth.String()),
			observability.Int64("total_cents", invoiceTotal.Cents()),
		)

		monthlyAggregate, err := u.repo.FindOrCreateMonthly(ctx, tx, userID, referenceMonth)
		if err != nil {
			return fmt.Errorf("failed to find or create monthly transaction: %w", err)
		}

		// Snapshot de IDs existentes antes da mutação do aggregate (determina INSERT vs UPDATE)
		existingIDs := make(map[string]struct{}, len(monthlyAggregate.Items))
		for _, item := range monthlyAggregate.Items {
			existingIDs[item.ID.String()] = struct{}{}
		}

		if err := monthlyAggregate.UpdateOrCreateCreditCardItem(categoryID, invoiceTotal, false); err != nil {
			return fmt.Errorf("failed to update or create credit card item: %w", err)
		}

		if err := u.ccItemPersister.Persist(ctx, tx, monthlyAggregate, existingIDs); err != nil {
			return fmt.Errorf("failed to persist credit card items: %w", err)
		}

		if err := u.repo.UpdateMonthly(ctx, tx, monthlyAggregate); err != nil {
			return fmt.Errorf("failed to update monthly transaction: %w", err)
		}

		return nil
	})

	if err != nil {
		u.o11y.Logger().Error(ctx, "failed to sync monthly from invoices", observability.Error(err))
		return err
	}

	return nil
}
