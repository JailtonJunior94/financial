package entities

import (
	"fmt"
	"testing"
	"time"

	"github.com/jailtonjunior94/financial/pkg/vos"
)

func TestNewBudgetItem(t *testing.T) {
	userID, err := vos.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	amount := vos.NewMoney(14_400.00)
	budget := NewBudget(userID, amount, time.Now().UTC())

	categoryID, err := vos.NewUUID()
	if err != nil {
		t.Fatal(err)
	}
	percentageGoal := vos.NewPercentage(30)

	type args struct {
		budget         *Budget
		categoryID     vos.UUID
		percentageGoal vos.Percentage
	}

	scenarios := []struct {
		name     string
		args     args
		expected func(budgetItem *BudgetItem, err error)
	}{{
		name: "should create a new budget item",
		args: args{
			budget:         budget,
			categoryID:     categoryID,
			percentageGoal: percentageGoal,
		},
		expected: func(budgetItem *BudgetItem, err error) {
			fmt.Println(budgetItem, err)
		},
	}}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			budgetItem := NewBudgetItem(
				scenario.args.budget,
				scenario.args.categoryID,
				scenario.args.percentageGoal,
			)
			scenario.expected(budgetItem, err)
		})
	}
}
