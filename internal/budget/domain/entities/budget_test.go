//go:build integration
// +build integration

package entities

import (
	"testing"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	budgetVos "github.com/jailtonjunior94/financial/internal/budget/domain/vos"

	"github.com/stretchr/testify/assert"
)

func TestNewBudget(t *testing.T) {
	userID, _ := vos.NewUUID()
	amount, _ := vos.NewMoneyFromFloat(14_400.00, vos.CurrencyBRL)

	type args struct {
		userID          vos.UUID
		amount          vos.Money
		referenceMonth budgetVos.ReferenceMonth
	}

	scenarios := []struct {
		name     string
		args     args
		expected func(budget *Budget)
	}{
		{
			name: "should create a new budget",
			args: args{
				userID:         userID,
				amount:         amount,
				referenceMonth: budgetVos.NewReferenceMonthFromDate(time.Now().UTC()),
			},
			expected: func(budget *Budget) {
				assert.NotNil(t, budget)
				expectedAmount, _ := vos.NewMoneyFromFloat(14_400.00, vos.CurrencyBRL)
				assert.True(t, budget.TotalAmount.Equals(expectedAmount))
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			budget := NewBudget(
				scenario.args.userID,
				scenario.args.amount,
				scenario.args.referenceMonth,
			)
			scenario.expected(budget)
		})
	}
}

func TestAddItems(t *testing.T) {
	userID, _ := vos.NewUUID()
	amount, _ := vos.NewMoneyFromFloat(14_400.00, vos.CurrencyBRL)
	referenceMonth := budgetVos.NewReferenceMonthFromDate(time.Now().UTC())
	budget := NewBudget(userID, amount, referenceMonth)

	categoryOne, _ := vos.NewUUID()
	percentageGoalCategoryOne, _ := vos.NewPercentage(30000) // 30% with scale 3

	categoryTwo, _ := vos.NewUUID()
	percentageGoalCategoryTwo, _ := vos.NewPercentage(15000) // 15%

	categoryThree, _ := vos.NewUUID()
	percentageGoalCategoryThree, _ := vos.NewPercentage(10000) // 10%

	categoryFour, _ := vos.NewUUID()
	percentageGoalCategoryFour, _ := vos.NewPercentage(10000) // 10%

	categoryFive, _ := vos.NewUUID()
	percentageGoalCategoryFive, _ := vos.NewPercentage(30000) // 30%

	categorySix, _ := vos.NewUUID()
	percentageGoalCategorySix, _ := vos.NewPercentage(5000) // 5%

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
		expected func(err error)
	}{
		{
			name: "should add items to the budget",
			args: args{
				budget: budget,
				items:  items,
			},
			expected: func(err error) {
				assert.Nil(t, err)
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			err := scenario.args.budget.AddItems(scenario.args.items)
			scenario.expected(err)
		})
	}
}
