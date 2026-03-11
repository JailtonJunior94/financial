package factories

import (
	"fmt"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
	pkgVos "github.com/jailtonjunior94/financial/pkg/domain/vos"
	"github.com/jailtonjunior94/financial/pkg/money"
)

// CreateBudgetParams holds the raw input for creating a budget.
type CreateBudgetParams struct {
	UserID         string
	ReferenceMonth string
	TotalAmount    string
	Currency       string
	Items          []CreateBudgetItemParams
}

// CreateBudgetItemParams holds the raw input for a budget item.
type CreateBudgetItemParams struct {
	CategoryID     string
	PercentageGoal string
}

func CreateBudget(userID string, params *CreateBudgetParams) (*entities.Budget, error) {
	// Parse user ID
	user, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("create_budget: invalid user ID: %w", err)
	}

	// Generate budget ID
	budgetID, err := vos.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create_budget: failed to generate budget ID: %w", err)
	}

	// Map currency string to Currency type
	var currency vos.Currency
	switch params.Currency {
	case "BRL":
		currency = vos.CurrencyBRL
	case "USD":
		currency = vos.CurrencyUSD
	case "EUR":
		currency = vos.CurrencyEUR
	default:
		return nil, fmt.Errorf("create_budget: unsupported currency: %s", params.Currency)
	}

	// Parse total amount from string (half-even rounding)
	totalAmount, err := money.NewMoney(params.TotalAmount, currency)
	if err != nil {
		return nil, fmt.Errorf("create_budget: invalid total amount: %w", err)
	}

	// Parse reference month
	referenceMonth, err := pkgVos.NewReferenceMonth(params.ReferenceMonth)
	if err != nil {
		return nil, fmt.Errorf("create_budget: %w", err)
	}

	// Create budget
	budget := entities.NewBudget(user, totalAmount, referenceMonth)
	budget.SetID(budgetID)

	// Create budget items
	var budgetItems []*entities.BudgetItem
	for _, itemInput := range params.Items {
		category, err := vos.NewUUIDFromString(itemInput.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("create_budget: invalid category ID: %w", err)
		}

		budgetItemID, err := vos.NewUUID()
		if err != nil {
			return nil, fmt.Errorf("create_budget: failed to generate item ID: %w", err)
		}

		// Parse percentage from string (e.g., "25.50" -> Percentage VO, half-even)
		percentage, err := money.NewPercentageFromString(itemInput.PercentageGoal)
		if err != nil {
			return nil, fmt.Errorf("create_budget: invalid percentage: %w", err)
		}

		newItem := entities.NewBudgetItem(budget.ID, budget.TotalAmount, category, percentage)
		newItem.SetID(budgetItemID)

		budgetItems = append(budgetItems, newItem)
	}

	// Add all items at once (validates 100% and prevents duplicates)
	if err := budget.AddItems(budgetItems); err != nil {
		return nil, fmt.Errorf("create_budget: %w", err)
	}

	return budget, nil
}
