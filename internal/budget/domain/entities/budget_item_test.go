//go:build integration
// +build integration

package entities

import (
	"testing"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"

	"github.com/stretchr/testify/assert"
)

func TestNewBudgetItem(t *testing.T) {
	userID, err := vos.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	amount, err := vos.NewMoneyFromFloat(14_400.00, vos.CurrencyBRL)
	if err != nil {
		t.Fatal(err)
	}

	referenceMonth := pkgVos.NewReferenceMonthFromDate(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	budget := NewBudget(userID, amount, referenceMonth)

	categoryID, err := vos.NewUUID()
	if err != nil {
		t.Fatal(err)
	}
	percentageGoal, err := vos.NewPercentage(30000) // 30% with scale 3
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		budget         *Budget
		categoryID     vos.UUID
		percentageGoal vos.Percentage
	}

	scenarios := []struct {
		name     string
		args     args
		expected func(budgetItem *BudgetItem, err error)
	}{
		{
			name: "should create a new budget item",
			args: args{
				budget:         budget,
				categoryID:     categoryID,
				percentageGoal: percentageGoal,
			},
			expected: func(budgetItem *BudgetItem, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, budgetItem)
				expectedAmount, _ := vos.NewMoneyFromFloat(4320.00, vos.CurrencyBRL)
				assert.True(t, budgetItem.PlannedAmount.Equals(expectedAmount))
			},
		},
	}

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

func TestUpdateItemSpentAmount(t *testing.T) {
	userID, err := vos.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	amount, err := vos.NewMoneyFromFloat(14_400.00, vos.CurrencyBRL)
	if err != nil {
		t.Fatal(err)
	}

	referenceMonth := pkgVos.NewReferenceMonthFromDate(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	budget := NewBudget(userID, amount, referenceMonth)

	categoryID, err := vos.NewUUID()
	if err != nil {
		t.Fatal(err)
	}
	percentageGoal, err := vos.NewPercentage(30000) // 30% with scale 3
	if err != nil {
		t.Fatal(err)
	}
	spentAmount, err := vos.NewMoneyFromFloat(1_323.80, vos.CurrencyBRL)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		budget         *Budget
		categoryID     vos.UUID
		percentageGoal vos.Percentage
		spentAmount    vos.Money
	}

	scenarios := []struct {
		name     string
		args     args
		expected func(budget *Budget, budgetItem *BudgetItem, err error)
	}{
		{
			name: "should update spent amount via aggregate",
			args: args{
				budget:         budget,
				categoryID:     categoryID,
				percentageGoal: percentageGoal,
				spentAmount:    spentAmount,
			},
			expected: func(budget *Budget, budgetItem *BudgetItem, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, budgetItem)
				expectedSpent, _ := vos.NewMoneyFromFloat(1_323.80, vos.CurrencyBRL)
				assert.True(t, budgetItem.SpentAmount.Equals(expectedSpent))
				expectedPlanned, _ := vos.NewMoneyFromFloat(4320.00, vos.CurrencyBRL)
				assert.True(t, budgetItem.PlannedAmount.Equals(expectedPlanned))
				// PercentageSpent = (SpentAmount / PlannedAmount) * 100
				// (1323.80 / 4320.00) * 100 = 30.64%
				expectedPercentage, _ := vos.NewPercentageFromFloat(30.64)
				actualPercentage := budgetItem.PercentageSpent()
				actualValue, _ := actualPercentage.Value()
				expectedValue, _ := expectedPercentage.Value()
				actualInt64 := actualValue.(int64)
				expectedInt64 := expectedValue.(int64)
				assert.True(t, actualInt64 >= expectedInt64-100 &&
					actualInt64 <= expectedInt64+100,
					"Expected percentage around %v, got %v", expectedInt64, actualInt64)
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			budgetItem := NewBudgetItem(
				scenario.args.budget,
				scenario.args.categoryID,
				scenario.args.percentageGoal,
			)

			// Add item to budget first
			err := scenario.args.budget.AddItem(budgetItem)
			assert.Nil(t, err)

			// Update spent amount via aggregate root
			err = scenario.args.budget.UpdateItemSpentAmount(budgetItem.ID, scenario.args.spentAmount)
			scenario.expected(scenario.args.budget, budgetItem, err)
		})
	}
}
