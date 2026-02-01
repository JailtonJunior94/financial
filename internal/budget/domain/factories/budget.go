package factories

import (
	"fmt"
	"strconv"

	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/budget/application/dtos"
	"github.com/jailtonjunior94/financial/internal/budget/domain/entities"
	budgetVos "github.com/jailtonjunior94/financial/internal/budget/domain/vos"
)

func CreateBudget(userID string, input *dtos.BudgetCreateInput) (*entities.Budget, error) {
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
	switch input.Currency {
	case "BRL":
		currency = vos.CurrencyBRL
	case "USD":
		currency = vos.CurrencyUSD
	case "EUR":
		currency = vos.CurrencyEUR
	default:
		return nil, fmt.Errorf("create_budget: unsupported currency: %s", input.Currency)
	}

	// Parse total amount from string (preserves precision)
	totalAmount, err := vos.NewMoneyFromString(input.TotalAmount, currency)
	if err != nil {
		return nil, fmt.Errorf("create_budget: invalid total amount: %w", err)
	}

	// Parse reference month
	referenceMonth, err := budgetVos.NewReferenceMonth(input.ReferenceMonth)
	if err != nil {
		return nil, fmt.Errorf("create_budget: %w", err)
	}

	// Create budget
	budget := entities.NewBudget(user, totalAmount, referenceMonth)
	budget.SetID(budgetID)

	// Create budget items
	var budgetItems []*entities.BudgetItem
	for _, itemInput := range input.Items {
		category, err := vos.NewUUIDFromString(itemInput.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("create_budget: invalid category ID: %w", err)
		}

		budgetItemID, err := vos.NewUUID()
		if err != nil {
			return nil, fmt.Errorf("create_budget: failed to generate item ID: %w", err)
		}

		// Parse percentage from string (e.g., "25.50" -> Percentage VO)
		percentage, err := parsePercentage(itemInput.PercentageGoal)
		if err != nil {
			return nil, fmt.Errorf("create_budget: invalid percentage: %w", err)
		}

		newItem := entities.NewBudgetItem(budget, category, percentage)
		newItem.SetID(budgetItemID)

		budgetItems = append(budgetItems, newItem)
	}

	// Add all items at once (validates 100% and prevents duplicates)
	if err := budget.AddItems(budgetItems); err != nil {
		return nil, fmt.Errorf("create_budget: %w", err)
	}

	return budget, nil
}

// parsePercentage converts a string percentage (e.g., "25.50") to a Percentage VO.
// This avoids float precision issues by parsing the string directly to scaled int64.
func parsePercentage(s string) (vos.Percentage, error) {
	// Parse string to float64
	percentageFloat, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return vos.Percentage{}, fmt.Errorf("invalid percentage format: %w", err)
	}

	// Convert to int64 with scale 3 (25.5% → 25500)
	// Using round to avoid truncation issues
	percentageInt := int64(percentageFloat*1000 + 0.5)

	return vos.NewPercentage(percentageInt)
}
