package entities

import (
	"testing"
	"time"

	"github.com/jailtonjunior94/financial/pkg/vos"

	"github.com/stretchr/testify/assert"
)

func TestNewBudget(t *testing.T) {
	userID, _ := vos.NewUUID()
	amount := vos.NewMoney(14_400.00)

	type args struct {
		userID vos.UUID
		amount vos.Money
		date   time.Time
	}

	scenarios := []struct {
		name     string
		args     args
		expected func(budget *Budget)
	}{{
		name: "should create a new budget",
		args: args{
			userID: userID,
			amount: amount,
			date:   time.Now().UTC(),
		},
		expected: func(budget *Budget) {
			assert.NotNil(t, budget)
			assert.True(t, budget.AmountGoal.Equals(vos.NewMoney(14_400.00)))
		},
	}}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			budget := NewBudget(
				scenario.args.userID,
				scenario.args.amount,
				scenario.args.date,
			)
			scenario.expected(budget)
		})
	}
}

func TestAddItems(t *testing.T) {
	userID, _ := vos.NewUUID()
	amount := vos.NewMoney(14_400.00)
	budget := NewBudget(userID, amount, time.Now().UTC())

	categoryOne, _ := vos.NewUUID()
	percentageGoalCategoryOne := vos.NewPercentage(30)

	categoryTwo, _ := vos.NewUUID()
	percentageGoalCategoryTwo := vos.NewPercentage(15)

	categoryThree, _ := vos.NewUUID()
	percentageGoalCategoryThree := vos.NewPercentage(10)

	categoryFour, _ := vos.NewUUID()
	percentageGoalCategoryFour := vos.NewPercentage(10)

	categoryFive, _ := vos.NewUUID()
	percentageGoalCategoryFive := vos.NewPercentage(30)

	categorySix, _ := vos.NewUUID()
	percentageGoalCategorySix := vos.NewPercentage(5)

	items := []*BudgetItem{
		NewBudgetItem(budget, categoryOne, percentageGoalCategoryOne),
		NewBudgetItem(budget, categoryTwo, percentageGoalCategoryTwo),
		NewBudgetItem(budget, categoryThree, percentageGoalCategoryThree),
		NewBudgetItem(budget, categoryFour, percentageGoalCategoryFour),
		NewBudgetItem(budget, categoryFive, percentageGoalCategoryFive),
		NewBudgetItem(budget, categorySix, percentageGoalCategorySix),
	}

	type args struct {
		budget *Budget
		items  []*BudgetItem
	}

	scenarios := []struct {
		name     string
		args     args
		expected func(sValid bool)
	}{{
		name: "should add items to the budget",
		args: args{
			budget: budget,
			items:  items,
		},
		expected: func(isValid bool) {
			assert.True(t, isValid)
		},
	}}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			isValid := scenario.args.budget.AddItems(scenario.args.items)
			scenario.expected(isValid)
		})
	}
}
