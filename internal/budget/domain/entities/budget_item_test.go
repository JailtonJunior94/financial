package entities

import (
	"testing"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/stretchr/testify/assert"
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
			assert.Nil(t, err)
			assert.NotNil(t, budgetItem)
			assert.True(t, budgetItem.AmountGoal.Equals(vos.NewMoney(432000)))
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

func TestAddAmountUsed(t *testing.T) {
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
	amountUsed := vos.NewMoney(1_323.80)

	type args struct {
		budget         *Budget
		categoryID     vos.UUID
		percentageGoal vos.Percentage
		amountUsed     vos.Money
	}

	scenarios := []struct {
		name     string
		args     args
		expected func(budgetItem *BudgetItem, err error)
	}{{
		name: "should calculate the amount goal",
		args: args{
			budget:         budget,
			categoryID:     categoryID,
			percentageGoal: percentageGoal,
			amountUsed:     amountUsed,
		},
		expected: func(budgetItem *BudgetItem, err error) {
			assert.Nil(t, err)
			assert.NotNil(t, budgetItem)
			assert.True(t, budgetItem.AmountUsed.Equals(vos.NewMoney(1_323.80)))
			assert.True(t, budgetItem.AmountGoal.Equals(vos.NewMoney(432000)))
			assert.True(t, budgetItem.PercentageUsed.Equals(vos.NewPercentage(30.00)))
		},
	}}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			budgetItem := NewBudgetItem(
				scenario.args.budget,
				scenario.args.categoryID,
				scenario.args.percentageGoal,
			)

			budgetItem.AddAmountUsed(scenario.args.amountUsed)
			scenario.expected(budgetItem, err)
		})
	}
}
