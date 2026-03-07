package events_test

import (
	"testing"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financial/internal/transaction/domain/events"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

func buildEvent(t *testing.T) *events.TransactionCreatedEvent {
	t.Helper()
	txID, _ := vos.NewUUID()
	userID, _ := vos.NewUUID()
	categoryID, _ := vos.NewUUID()
	amount, _ := vos.NewMoneyFromFloat(100.00, vos.CurrencyBRL)
	pm, _ := transactionVos.NewPaymentMethod(transactionVos.PaymentMethodPix)
	refMonth, _ := pkgVos.NewReferenceMonth("2026-03")
	return events.NewTransactionCreatedEvent(
		txID,
		userID,
		categoryID,
		amount,
		pm,
		time.Now().Add(-time.Hour),
		refMonth,
		nil,
		nil,
		nil,
		nil,
	)
}

func TestTransactionCreatedEvent(t *testing.T) {
	t.Run("EventType should return transaction.created", func(t *testing.T) {
		e := buildEvent(t)
		require.Equal(t, "transaction.created", e.EventType())
	})

	t.Run("IdempotencyKey should return transaction_id as string", func(t *testing.T) {
		txID, _ := vos.NewUUID()
		userID, _ := vos.NewUUID()
		categoryID, _ := vos.NewUUID()
		amount, _ := vos.NewMoneyFromFloat(50.00, vos.CurrencyBRL)
		pm, _ := transactionVos.NewPaymentMethod(transactionVos.PaymentMethodPix)
		refMonth, _ := pkgVos.NewReferenceMonth("2026-03")
		e := events.NewTransactionCreatedEvent(txID, userID, categoryID, amount, pm, time.Now().Add(-time.Hour), refMonth, nil, nil, nil, nil)
		require.Equal(t, txID.String(), e.IdempotencyKey())
	})

	t.Run("Payload should contain required fields", func(t *testing.T) {
		e := buildEvent(t)
		payload := e.Payload()
		require.Contains(t, payload, "version")
		require.Contains(t, payload, "transaction_id")
		require.Contains(t, payload, "user_id")
		require.Contains(t, payload, "amount")
		require.Contains(t, payload, "reference_month")
		require.Equal(t, "2026-03", payload["reference_month"])
	})

	t.Run("Payload with nil invoice_id should have null invoice_id field", func(t *testing.T) {
		e := buildEvent(t)
		payload := e.Payload()
		require.Nil(t, payload["invoice_id"])
	})

	t.Run("Payload with installment_group_id should include it correctly", func(t *testing.T) {
		txID, _ := vos.NewUUID()
		userID, _ := vos.NewUUID()
		categoryID, _ := vos.NewUUID()
		groupID, _ := vos.NewUUID()
		amount, _ := vos.NewMoneyFromFloat(100.00, vos.CurrencyBRL)
		pm, _ := transactionVos.NewPaymentMethod(transactionVos.PaymentMethodCredit)
		refMonth, _ := pkgVos.NewReferenceMonth("2026-03")
		num := 1
		total := 3
		e := events.NewTransactionCreatedEvent(txID, userID, categoryID, amount, pm, time.Now().Add(-time.Hour), refMonth, nil, &num, &total, &groupID)
		payload := e.Payload()
		require.Equal(t, groupID.String(), payload["installment_group_id"])
	})
}
