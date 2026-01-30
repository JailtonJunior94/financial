package usecase

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/category/application/dtos"
	"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/pagination"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type (
	// FindCategoryPaginatedUseCase lista categorias de um usuário com paginação cursor-based.
	FindCategoryPaginatedUseCase interface {
		Execute(ctx context.Context, input FindCategoryPaginatedInput) (*FindCategoryPaginatedOutput, error)
	}

	// FindCategoryPaginatedInput representa a entrada do use case.
	FindCategoryPaginatedInput struct {
		UserID string
		Limit  int
		Cursor string
	}

	// FindCategoryPaginatedOutput representa a saída do use case.
	FindCategoryPaginatedOutput struct {
		Categories []*dtos.CategoryOutput
		NextCursor *string
	}

	findCategoryPaginatedUseCase struct {
		o11y       observability.Observability
		repository interfaces.CategoryRepository
	}
)

// NewFindCategoryPaginatedUseCase cria uma nova instância do use case.
func NewFindCategoryPaginatedUseCase(
	o11y observability.Observability,
	repository interfaces.CategoryRepository,
) FindCategoryPaginatedUseCase {
	return &findCategoryPaginatedUseCase{
		o11y:       o11y,
		repository: repository,
	}
}

// Execute executa o use case de listagem paginada de categorias.
func (u *findCategoryPaginatedUseCase) Execute(
	ctx context.Context,
	input FindCategoryPaginatedInput,
) (*FindCategoryPaginatedOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "find_category_paginated_usecase.execute")
	defer span.End()

	// Parse user ID
	userID, err := vos.NewUUIDFromString(input.UserID)
	if err != nil {
		span.AddEvent(
			"error parsing user id",
			observability.Field{Key: "user_id", Value: input.UserID},
			observability.Field{Key: "error", Value: err},
		)

		return nil, err
	}

	// Decode cursor
	cursor, err := pagination.DecodeCursor(input.Cursor)
	if err != nil {
		span.AddEvent(
			"error decoding cursor",
			observability.Field{Key: "cursor", Value: input.Cursor},
			observability.Field{Key: "error", Value: err},
		)

		return nil, err
	}

	// List categories (paginado)
	categories, err := u.repository.ListPaginated(ctx, interfaces.ListCategoriesParams{
		UserID: userID,
		Limit:  input.Limit + 1, // +1 para detectar has_next
		Cursor: cursor,
	})
	if err != nil {
		span.AddEvent(
			"error listing categories from repository",
			observability.Field{Key: "user_id", Value: input.UserID},
			observability.Field{Key: "error", Value: err},
		)

		return nil, err
	}

	// Determinar se há próxima página
	hasNext := len(categories) > input.Limit
	if hasNext {
		categories = categories[:input.Limit] // Remover o item extra
	}

	// Construir cursor para próxima página
	var nextCursor *string
	if hasNext && len(categories) > 0 {
		lastCategory := categories[len(categories)-1]

		newCursor := pagination.Cursor{
			Fields: map[string]interface{}{
				"sequence": lastCategory.Sequence.Value(),
				"id":       lastCategory.ID.String(),
			},
		}

		encoded, err := pagination.EncodeCursor(newCursor)
		if err != nil {
			span.AddEvent(
				"error encoding cursor",
				observability.Field{Key: "error", Value: err},
			)

			return nil, err
		}

		nextCursor = &encoded
	}

	// Converter para DTOs
	output := make([]*dtos.CategoryOutput, len(categories))
	for i, category := range categories {
		output[i] = &dtos.CategoryOutput{
			ID:        category.ID.String(),
			Name:      category.Name.String(),
			Sequence:  category.Sequence.Value(),
			CreatedAt: category.CreatedAt.ValueOr(time.Time{}),
		}
	}

	return &FindCategoryPaginatedOutput{
		Categories: output,
		NextCursor: nextCursor,
	}, nil
}
