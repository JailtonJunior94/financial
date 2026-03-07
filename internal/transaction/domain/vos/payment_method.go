package vos

import "github.com/jailtonjunior94/financial/internal/transaction/domain"

const (
	PaymentMethodPix    = "pix"
	PaymentMethodBoleto = "boleto"
	PaymentMethodTed    = "ted"
	PaymentMethodDebit  = "debit"
	PaymentMethodCredit = "credit"
)

type PaymentMethod struct {
	Value string
}

func NewPaymentMethod(v string) (PaymentMethod, error) {
	switch v {
	case PaymentMethodPix, PaymentMethodBoleto, PaymentMethodTed, PaymentMethodDebit, PaymentMethodCredit:
		return PaymentMethod{Value: v}, nil
	default:
		return PaymentMethod{}, domain.ErrInvalidPaymentMethod
	}
}

func (p PaymentMethod) IsCredit() bool {
	return p.Value == PaymentMethodCredit
}

func (p PaymentMethod) RequiresCard() bool {
	return p.Value == PaymentMethodCredit || p.Value == PaymentMethodDebit
}

func (p PaymentMethod) String() string {
	return p.Value
}
