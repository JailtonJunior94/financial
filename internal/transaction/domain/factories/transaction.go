package factories

import (
	"fmt"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	transactionDomain "github.com/jailtonjunior94/financial/internal/transaction/domain"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
)

// CreateParams holds the raw input for creating a single transaction.
type CreateParams struct {
	UserID          string
	CategoryID      string
	SubcategoryID   string
	CardID          string
	InvoiceID       string
	Description     string
	Amount          float64
	PaymentMethod   string
	TransactionDate time.Time
	Installments    int
}

// InstallmentParams holds the raw input for creating installment transactions.
type InstallmentParams struct {
	CreateParams
	InvoiceIDs []string
}

// TransactionFactory creates Transaction entities from raw input.
type TransactionFactory struct{}

// NewTransactionFactory returns a new TransactionFactory.
func NewTransactionFactory() *TransactionFactory {
	return &TransactionFactory{}
}

// Create validates input and creates a single Transaction.
func (f *TransactionFactory) Create(params CreateParams) (*entities.Transaction, error) {
	pm, err := transactionVos.NewPaymentMethod(params.PaymentMethod)
	if err != nil {
		return nil, err
	}
	if pm.RequiresCard() && params.CardID == "" {
		return nil, transactionDomain.ErrCardRequiredForCredit
	}
	if !pm.RequiresCard() && params.CardID != "" {
		return nil, transactionDomain.ErrCardNotAllowedForMethod
	}
	userID, err := vos.NewUUIDFromString(params.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}
	categoryID, err := vos.NewUUIDFromString(params.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid category_id: %w", err)
	}
	txID, err := vos.NewUUID()
	if err != nil {
		return nil, err
	}
	amount, err := vos.NewMoneyFromFloat(params.Amount, vos.CurrencyBRL)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}
	subcategoryID, err := parseOptionalUUID(params.SubcategoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid subcategory_id: %w", err)
	}
	cardID, err := parseOptionalUUID(params.CardID)
	if err != nil {
		return nil, fmt.Errorf("invalid card_id: %w", err)
	}
	invoiceID, err := parseOptionalUUID(params.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice_id: %w", err)
	}
	installments := params.Installments
	if installments <= 0 {
		installments = 1
	}
	installmentNumber := 1
	status, err := transactionVos.NewTransactionStatus(transactionVos.TransactionStatusActive)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction status: %w", err)
	}
	return entities.NewTransaction(entities.TransactionParams{
		ID:                txID,
		UserID:            userID,
		CategoryID:        categoryID,
		SubcategoryID:     subcategoryID,
		CardID:            cardID,
		InvoiceID:         invoiceID,
		Description:       params.Description,
		Amount:            amount,
		PaymentMethod:     pm,
		TransactionDate:   params.TransactionDate,
		InstallmentNumber: &installmentNumber,
		InstallmentTotal:  &installments,
		Status:            status,
		CreatedAt:         time.Now().UTC(),
	})
}

// CreateInstallments generates N installment transactions with a shared group ID.
// The last installment absorbs any rounding difference.
func (f *TransactionFactory) CreateInstallments(params InstallmentParams) ([]*entities.Transaction, error) {
	n := params.Installments
	if n <= 0 {
		n = 1
	}
	if len(params.InvoiceIDs) != n {
		return nil, fmt.Errorf("invoice_ids length (%d) must equal installments (%d)", len(params.InvoiceIDs), n)
	}
	pm, err := transactionVos.NewPaymentMethod(params.PaymentMethod)
	if err != nil {
		return nil, err
	}
	if !pm.IsCredit() {
		return nil, transactionDomain.ErrInstallmentsOnlyForCredit
	}
	if params.CardID == "" {
		return nil, transactionDomain.ErrCardRequiredForCredit
	}
	userID, err := vos.NewUUIDFromString(params.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}
	categoryID, err := vos.NewUUIDFromString(params.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid category_id: %w", err)
	}
	cardUUID, err := vos.NewUUIDFromString(params.CardID)
	if err != nil {
		return nil, fmt.Errorf("invalid card_id: %w", err)
	}
	cardID := &cardUUID
	totalAmount, err := vos.NewMoneyFromFloat(params.Amount, vos.CurrencyBRL)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}
	groupID, err := vos.NewUUID()
	if err != nil {
		return nil, err
	}
	subcategoryID, err := parseOptionalUUID(params.SubcategoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid subcategory_id: %w", err)
	}
	status, err := transactionVos.NewTransactionStatus(transactionVos.TransactionStatusActive)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction status: %w", err)
	}
	totalCents := totalAmount.Cents()
	perInstallmentCents := totalCents / int64(n)
	transactions := make([]*entities.Transaction, 0, n)
	var sumCents int64
	for i := 0; i < n; i++ {
		installmentNumber := i + 1
		var installmentCents int64
		if i < n-1 {
			installmentCents = perInstallmentCents
			sumCents += installmentCents
		} else {
			installmentCents = totalCents - sumCents
		}
		installmentAmount, err := vos.NewMoney(installmentCents, vos.CurrencyBRL)
		if err != nil {
			return nil, err
		}
		invoiceUUID, err := vos.NewUUIDFromString(params.InvoiceIDs[i])
		if err != nil {
			return nil, fmt.Errorf("invalid invoice_id[%d]: %w", i, err)
		}
		invoiceID := &invoiceUUID
		txID, err := vos.NewUUID()
		if err != nil {
			return nil, err
		}
		tx, err := entities.NewTransaction(entities.TransactionParams{
			ID:                 txID,
			UserID:             userID,
			CategoryID:         categoryID,
			SubcategoryID:      subcategoryID,
			CardID:             cardID,
			InvoiceID:          invoiceID,
			InstallmentGroupID: &groupID,
			Description:        params.Description,
			Amount:             installmentAmount,
			PaymentMethod:      pm,
			TransactionDate:    params.TransactionDate,
			InstallmentNumber:  &installmentNumber,
			InstallmentTotal:   &n,
			Status:             status,
			CreatedAt:          time.Now().UTC(),
		})
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}
	return transactions, nil
}

func parseOptionalUUID(value string) (*vos.UUID, error) {
	if value == "" {
		return nil, nil
	}
	uid, err := vos.NewUUIDFromString(value)
	if err != nil {
		return nil, err
	}
	return &uid, nil
}
