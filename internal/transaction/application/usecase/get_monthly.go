package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"

	"github.com/jailtonjunior94/financial/internal/transaction/application/dtos"
	"github.com/jailtonjunior94/financial/internal/transaction/domain/interfaces"
)

type (
	// GetMonthlyUseCase busca uma monthly transaction por ID.
	GetMonthlyUseCase interface {
		Execute(ctx context.Context, userID string, monthlyID string) (*dtos.MonthlyTransactionOutput, error)
	}

	getMonthlyUseCase struct {
		o11y       observability.Observability
		repository interfaces.TransactionRepository
	}
)

// NewGetMonthlyUseCase cria uma nova instância do use case.
func NewGetMonthlyUseCase(
	o11y observability.Observability,
	repository interfaces.TransactionRepository,
) GetMonthlyUseCase {
	return &getMonthlyUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

// Execute executa o use case de busca de monthly transaction.
func (u *getMonthlyUseCase) Execute(
	ctx context.Context,
	userID string,
	monthlyID string,
) (*dtos.MonthlyTransactionOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "get_monthly_usecase.execute")
	defer span.End()

	// Parse user ID
	userUUID, err := vos.NewUUIDFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	// Parse monthly ID
	monthlyUUID, err := vos.NewUUIDFromString(monthlyID)
	if err != nil {
		return nil, fmt.Errorf("invalid monthly_id: %w", err)
	}

	// Get monthly transaction
	monthly, err := u.repository.GetMonthlyByID(ctx, userUUID, monthlyUUID)
	if err != nil {
		return nil, err
	}

	if monthly == nil {
		return nil, fmt.Errorf("monthly transaction not found")
	}

	// Converter items para DTOs (apenas itens não deletados)
	items := make([]*dtos.TransactionItemOutput, 0)
	for _, item := range monthly.Items {
		if item.IsDeleted() {
			continue
		}

		items = append(items, &dtos.TransactionItemOutput{
			ID:          item.ID.String(),
			CategoryID:  item.CategoryID.String(),
			Title:       item.Title,
			Description: item.Description,
			Amount:      fmt.Sprintf("%.2f", item.Amount.Float()),
			Direction:   item.Direction.String(),
			Type:        item.Type.String(),
			IsPaid:      item.IsPaid,
			CreatedAt:   item.CreatedAt.ValueOr(time.Time{}),
			UpdatedAt:   item.UpdatedAt.ValueOr(time.Time{}),
		})
	}

	return &dtos.MonthlyTransactionOutput{
		ID:             monthly.ID.String(),
		ReferenceMonth: monthly.ReferenceMonth.String(),
		TotalIncome:    fmt.Sprintf("%.2f", monthly.TotalIncome.Float()),
		TotalExpense:   fmt.Sprintf("%.2f", monthly.TotalExpense.Float()),
		TotalAmount:    fmt.Sprintf("%.2f", monthly.TotalAmount.Float()),
		Items:          items,
		CreatedAt:      monthly.CreatedAt.ValueOr(time.Time{}),
		UpdatedAt:      monthly.UpdatedAt.ValueOr(time.Time{}),
	}, nil
}
