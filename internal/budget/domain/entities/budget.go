package entities

import (
	"time"

	"github.com/jailtonjunior94/financial/pkg/entity"
	"github.com/jailtonjunior94/financial/pkg/vos"
)

type Budget struct {
	entity.Base
	UserID vos.UUID
	Date   time.Time
	Amount vos.Money
	Items  []*BudgetItem
}

func NewBudget(userID vos.UUID, amount vos.Money, date time.Time) *Budget {
	return &Budget{
		UserID: userID,
		Amount: amount,
		Date:   date,
		Base: entity.Base{
			CreatedAt: time.Now().UTC(),
		},
	}
}
