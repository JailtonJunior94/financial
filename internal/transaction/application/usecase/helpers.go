package usecase

import (
	"time"

	"github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/entities"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
)

func toOutput(t *entities.Transaction) *dtos.TransactionOutput {
	out := &dtos.TransactionOutput{
		ID:              t.ID.String(),
		UserID:          t.UserID.String(),
		CategoryID:      t.CategoryID.String(),
		Description:     t.Description,
		Amount:          t.Amount.Float(),
		PaymentMethod:   t.PaymentMethod.String(),
		TransactionDate: t.TransactionDate.Format("2006-01-02"),
		Status:          t.Status.String(),
		CreatedAt:       t.CreatedAt.Format(time.RFC3339),
	}
	if t.SubcategoryID != nil {
		s := t.SubcategoryID.String()
		out.SubcategoryID = &s
	}
	if t.CardID != nil {
		c := t.CardID.String()
		out.CardID = &c
	}
	if t.InvoiceID != nil {
		inv := t.InvoiceID.String()
		out.InvoiceID = &inv
	}
	if t.InstallmentGroupID != nil {
		g := t.InstallmentGroupID.String()
		out.InstallmentGroupID = &g
	}
	out.InstallmentNumber = t.InstallmentNumber
	out.InstallmentTotal = t.InstallmentTotal
	return out
}

func toOutputList(ts []*entities.Transaction) []*dtos.TransactionOutput {
	out := make([]*dtos.TransactionOutput, 0, len(ts))
	for _, t := range ts {
		out = append(out, toOutput(t))
	}
	return out
}

func resolveReferenceMonth(_ *entities.Transaction, transactionDate time.Time) pkgVos.ReferenceMonth {
	return pkgVos.NewReferenceMonthFromDate(transactionDate)
}
