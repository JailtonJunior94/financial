package entities

import (
	"fmt"
	"strings"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	transactionDomain "github.com/jailtonjunior94/financial/internal/transaction/domain"
	transactionVos "github.com/jailtonjunior94/financial/internal/transaction/domain/vos"
)

// TransactionParams holds all fields needed to create a Transaction.
type TransactionParams struct {
	ID                 vos.UUID
	UserID             vos.UUID
	CategoryID         vos.UUID
	SubcategoryID      *vos.UUID
	CardID             *vos.UUID
	InvoiceID          *vos.UUID
	InstallmentGroupID *vos.UUID
	Description        string
	Amount             vos.Money
	PaymentMethod      transactionVos.PaymentMethod
	TransactionDate    time.Time
	InstallmentNumber  *int
	InstallmentTotal   *int
	Status             transactionVos.TransactionStatus
	CreatedAt          time.Time
	UpdatedAt          *time.Time
	DeletedAt          *time.Time
}

// Transaction represents an individual financial transaction.
type Transaction struct {
	ID                 vos.UUID
	UserID             vos.UUID
	CategoryID         vos.UUID
	SubcategoryID      *vos.UUID
	CardID             *vos.UUID
	InvoiceID          *vos.UUID
	InstallmentGroupID *vos.UUID
	Description        string
	Amount             vos.Money
	PaymentMethod      transactionVos.PaymentMethod
	TransactionDate    time.Time
	InstallmentNumber  *int
	InstallmentTotal   *int
	Status             transactionVos.TransactionStatus
	CreatedAt          time.Time
	UpdatedAt          *time.Time
	DeletedAt          *time.Time
}

// NewTransaction creates a Transaction, validating description and amount.
func NewTransaction(params TransactionParams) (*Transaction, error) {
	if strings.TrimSpace(params.Description) == "" {
		return nil, fmt.Errorf("%w", transactionDomain.ErrDescriptionRequired)
	}
	if !params.Amount.IsPositive() {
		return nil, fmt.Errorf("%w", transactionDomain.ErrAmountMustBePositive)
	}
	return &Transaction{
		ID:                 params.ID,
		UserID:             params.UserID,
		CategoryID:         params.CategoryID,
		SubcategoryID:      params.SubcategoryID,
		CardID:             params.CardID,
		InvoiceID:          params.InvoiceID,
		InstallmentGroupID: params.InstallmentGroupID,
		Description:        params.Description,
		Amount:             params.Amount,
		PaymentMethod:      params.PaymentMethod,
		TransactionDate:    params.TransactionDate,
		InstallmentNumber:  params.InstallmentNumber,
		InstallmentTotal:   params.InstallmentTotal,
		Status:             params.Status,
		CreatedAt:          params.CreatedAt,
		UpdatedAt:          params.UpdatedAt,
		DeletedAt:          params.DeletedAt,
	}, nil
}

// Cancel sets the transaction status to cancelled.
func (t *Transaction) Cancel() {
	status, _ := transactionVos.NewTransactionStatus(transactionVos.TransactionStatusCancelled)
	t.Status = status
	now := time.Now().UTC()
	t.UpdatedAt = &now
}

// UpdateDetails updates the transaction's mutable fields.
func (t *Transaction) UpdateDetails(description string, amount vos.Money, categoryID vos.UUID) error {
	if !amount.IsPositive() {
		return fmt.Errorf("%w", transactionDomain.ErrAmountMustBePositive)
	}
	t.Description = description
	t.Amount = amount
	t.CategoryID = categoryID
	now := time.Now().UTC()
	t.UpdatedAt = &now
	return nil
}

// IsEditable returns false when the associated invoice is closed or paid.
func (t *Transaction) IsEditable(invoiceStatus string) bool {
	return invoiceStatus != "closed" && invoiceStatus != "paid"
}
