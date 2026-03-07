package domain

import (
	"net/http"

	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
)

func ErrorMappings() map[error]httperrors.ErrorMapping {
	return map[error]httperrors.ErrorMapping{
		ErrTransactionNotFound:       {Status: http.StatusNotFound, Message: "Transaction not found"},
		ErrTransactionNotOwned:       {Status: http.StatusForbidden, Message: "Access denied"},
		ErrInvoiceClosed:             {Status: http.StatusUnprocessableEntity, Message: "Invoice is closed"},
		ErrNothingToReverse:          {Status: http.StatusConflict, Message: "Nothing to reverse"},
		ErrCardRequiredForCredit:     {Status: http.StatusBadRequest, Message: "Card is required for this payment method"},
		ErrCardNotAllowedForMethod:   {Status: http.StatusBadRequest, Message: "Card not allowed for this payment method"},
		ErrInstallmentsOnlyForCredit: {Status: http.StatusBadRequest, Message: "Installments only allowed for credit"},
		ErrTransactionDateFuture:     {Status: http.StatusBadRequest, Message: "Transaction date cannot be in the future"},
		ErrInvalidPaymentMethod:      {Status: http.StatusBadRequest, Message: "Invalid payment method"},
		ErrInvalidTransactionStatus:  {Status: http.StatusBadRequest, Message: "Invalid transaction status"},
		ErrDescriptionRequired:       {Status: http.StatusBadRequest, Message: "Description is required"},
		ErrAmountMustBePositive:      {Status: http.StatusBadRequest, Message: "Amount must be positive"},
		ErrInstallmentsTooMany:       {Status: http.StatusBadRequest, Message: "Installments cannot exceed 48"},
	}
}
