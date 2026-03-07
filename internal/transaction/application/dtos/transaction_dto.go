package dtos

import (
	"fmt"
	"strings"
	"time"

	transactionDomain "github.com/jailtonjunior94/financial/internal/transaction/domain"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
)

// TransactionInput is the request body for POST /api/v1/transactions.
type TransactionInput struct {
	Description     string  `json:"description"`
	Amount          float64 `json:"amount"`
	PaymentMethod   string  `json:"payment_method"`
	TransactionDate string  `json:"transaction_date"`
	CategoryID      string  `json:"category_id"`
	SubcategoryID   string  `json:"subcategory_id,omitempty"`
	CardID          string  `json:"card_id,omitempty"`
	Installments    int     `json:"installments,omitempty"`
}

// Validate validates the TransactionInput fields.
func (i *TransactionInput) Validate() error {
	if strings.TrimSpace(i.Description) == "" {
		return transactionDomain.ErrDescriptionRequired
	}
	if i.Amount <= 0 {
		return transactionDomain.ErrAmountMustBePositive
	}
	pm, err := transactionVos.NewPaymentMethod(i.PaymentMethod)
	if err != nil {
		return transactionDomain.ErrInvalidPaymentMethod
	}
	if i.TransactionDate == "" {
		return fmt.Errorf("transaction_date is required")
	}
	parsed, err := time.Parse("2006-01-02", i.TransactionDate)
	if err != nil {
		return fmt.Errorf("transaction_date must be in YYYY-MM-DD format")
	}
	if parsed.After(time.Now().UTC().Truncate(24 * time.Hour)) {
		return transactionDomain.ErrTransactionDateFuture
	}
	if i.CategoryID == "" {
		return fmt.Errorf("category_id is required")
	}
	if pm.RequiresCard() && strings.TrimSpace(i.CardID) == "" {
		return transactionDomain.ErrCardRequiredForCredit
	}
	if !pm.RequiresCard() && strings.TrimSpace(i.CardID) != "" {
		return transactionDomain.ErrCardNotAllowedForMethod
	}
	installments := i.Installments
	if installments == 0 {
		installments = 1
	}
	if !pm.IsCredit() && installments > 1 {
		return transactionDomain.ErrInstallmentsOnlyForCredit
	}
	if pm.IsCredit() && installments > 48 {
		return transactionDomain.ErrInstallmentsTooMany
	}
	return nil
}

// TransactionUpdateInput is the request body for PUT /api/v1/transactions/{id}.
type TransactionUpdateInput struct {
	Description   string  `json:"description"`
	Amount        float64 `json:"amount"`
	CategoryID    string  `json:"category_id"`
	SubcategoryID string  `json:"subcategory_id,omitempty"`
}

// Validate validates the TransactionUpdateInput fields.
func (i *TransactionUpdateInput) Validate() error {
	if strings.TrimSpace(i.Description) == "" {
		return transactionDomain.ErrDescriptionRequired
	}
	if i.Amount <= 0 {
		return transactionDomain.ErrAmountMustBePositive
	}
	if strings.TrimSpace(i.CategoryID) == "" {
		return fmt.Errorf("category_id is required")
	}
	return nil
}

// TransactionOutput is the response for creation and retrieval.
type TransactionOutput struct {
	ID                 string  `json:"id"`
	UserID             string  `json:"user_id"`
	CategoryID         string  `json:"category_id"`
	SubcategoryID      *string `json:"subcategory_id,omitempty"`
	CardID             *string `json:"card_id,omitempty"`
	InvoiceID          *string `json:"invoice_id,omitempty"`
	InstallmentGroupID *string `json:"installment_group_id,omitempty"`
	Description        string  `json:"description"`
	Amount             float64 `json:"amount"`
	PaymentMethod      string  `json:"payment_method"`
	TransactionDate    string  `json:"transaction_date"`
	InstallmentNumber  *int    `json:"installment_number,omitempty"`
	InstallmentTotal   *int    `json:"installment_total,omitempty"`
	Status             string  `json:"status"`
	CreatedAt          string  `json:"created_at"`
}

// TransactionListOutput is the paginated response for GET /api/v1/transactions.
type TransactionListOutput struct {
	Data       []*TransactionOutput `json:"data"`
	NextCursor string               `json:"next_cursor,omitempty"`
}

// ReverseOutput is the response for POST /api/v1/transactions/{id}/reverse.
type ReverseOutput struct {
	Cancelled []*TransactionOutput `json:"cancelled"`
	Kept      []*TransactionOutput `json:"kept"`
}

// ListParams holds the filtering and pagination parameters for listing transactions.
type ListParams struct {
	PaymentMethod string
	CategoryID    string
	StartDate     string
	EndDate       string
	Limit         int
	Cursor        string
}
