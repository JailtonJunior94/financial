package entities

import (
	"time"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/vos"
)

type PaymentMethod struct {
	ID          sharedVos.UUID
	Name        vos.PaymentMethodName
	Code        vos.PaymentMethodCode
	Description vos.Description
	CreatedAt   sharedVos.NullableTime
	UpdatedAt   sharedVos.NullableTime
	DeletedAt   sharedVos.NullableTime
}

func NewPaymentMethod(name vos.PaymentMethodName, code vos.PaymentMethodCode, description vos.Description) (*PaymentMethod, error) {
	paymentMethod := &PaymentMethod{
		Name:        name,
		Code:        code,
		Description: description,
		CreatedAt:   sharedVos.NewNullableTime(time.Now()),
	}
	return paymentMethod, nil
}

func (p *PaymentMethod) Update(name string, description string) error {
	paymentMethodName, err := vos.NewPaymentMethodName(name)
	if err != nil {
		return err
	}

	paymentMethodDescription, err := vos.NewDescription(description)
	if err != nil {
		return err
	}

	p.Name = paymentMethodName
	p.Description = paymentMethodDescription
	p.UpdatedAt = sharedVos.NewNullableTime(time.Now())

	return nil
}

func (p *PaymentMethod) Delete() *PaymentMethod {
	p.DeletedAt = sharedVos.NewNullableTime(time.Now())
	return p
}
