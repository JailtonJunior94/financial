package factories

import (
	"fmt"

	"github.com/jailtonjunior94/financial/internal/payment_method/domain/entities"

	sharedVos "github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/jailtonjunior94/financial/internal/payment_method/domain/vos"
)

func CreatePaymentMethod(name, code, description string) (*entities.PaymentMethod, error) {
	id, err := sharedVos.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("error generating payment method id: %v", err)
	}

	paymentMethodName, err := vos.NewPaymentMethodName(name)
	if err != nil {
		return nil, err
	}

	paymentMethodCode, err := vos.NewPaymentMethodCode(code)
	if err != nil {
		return nil, err
	}

	paymentMethodDescription, err := vos.NewDescription(description)
	if err != nil {
		return nil, err
	}

	paymentMethod, err := entities.NewPaymentMethod(paymentMethodName, paymentMethodCode, paymentMethodDescription)
	if err != nil {
		return nil, fmt.Errorf("error creating payment method: %w", err)
	}

	paymentMethod.ID = id
	return paymentMethod, nil
}
