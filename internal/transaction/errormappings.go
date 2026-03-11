package transaction

import (
	"net/http"

	"github.com/jailtonjunior94/financial/internal/transaction/domain"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
)

// ErrorMappings returns the HTTP status mappings for transaction domain errors.
func ErrorMappings() map[error]httperrors.ErrorMapping {
	return map[error]httperrors.ErrorMapping{
		domain.ErrTransactionNotFound:       {Status: http.StatusNotFound, Message: "Transaction not found"},
		domain.ErrTransactionNotOwned:       {Status: http.StatusForbidden, Message: "Access denied"},
		domain.ErrInvoiceClosed:             {Status: http.StatusUnprocessableEntity, Message: "Invoice is closed"},
		domain.ErrNothingToReverse:          {Status: http.StatusConflict, Message: "Nothing to reverse"},
		domain.ErrCardRequiredForCredit:     {Status: http.StatusBadRequest, Message: "Card is required for this payment method"},
		domain.ErrCardNotAllowedForMethod:   {Status: http.StatusBadRequest, Message: "Card not allowed for this payment method"},
		domain.ErrInstallmentsOnlyForCredit: {Status: http.StatusBadRequest, Message: "Installments only allowed for credit"},
		domain.ErrTransactionDateFuture:     {Status: http.StatusBadRequest, Message: "Transaction date cannot be in the future"},
		domain.ErrInvalidPaymentMethod:      {Status: http.StatusBadRequest, Message: "Invalid payment method"},
		domain.ErrInvalidTransactionStatus:  {Status: http.StatusBadRequest, Message: "Invalid transaction status"},
		domain.ErrDescriptionRequired:       {Status: http.StatusBadRequest, Message: "Description is required"},
		domain.ErrAmountMustBePositive:      {Status: http.StatusBadRequest, Message: "Amount must be positive"},
		domain.ErrInstallmentsTooMany:       {Status: http.StatusBadRequest, Message: "Installments cannot exceed 48"},
	}
}
