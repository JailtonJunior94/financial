package adapters

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"

	pkginterfaces "github.com/jailtonjunior94/financial/pkg/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

type categoryProviderAdapter struct {
	db   database.DBTX
	o11y observability.Observability
	fm   *metrics.FinancialMetrics
}

func NewCategoryProviderAdapter(
	db database.DBTX,
	o11y observability.Observability,
	fm *metrics.FinancialMetrics,
) pkginterfaces.CategoryProvider {
	return &categoryProviderAdapter{db: db, o11y: o11y, fm: fm}
}

func (a *categoryProviderAdapter) ValidateCategories(ctx context.Context, userID string, categoryIDs []string) error {
	start := time.Now()
	ctx, span := a.o11y.Tracer().Start(ctx, "category_provider_adapter.validate_categories")
	defer span.End()

	if len(categoryIDs) == 0 {
		return nil
	}

	a.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "ValidateCategories"),
		observability.String("layer", "adapter"),
		observability.String("entity", "category"),
		observability.String("user_id", userID),
	)

	foundIDs, err := a.queryFoundIDs(ctx, userID, categoryIDs)
	if err != nil {
		span.RecordError(err)
		a.logQueryFailed(ctx, userID, err)
		a.fm.RecordRepositoryFailure(ctx, "validate_categories", "category", "infra", time.Since(start))
		return fmt.Errorf("category_provider_adapter.validate_categories: %w", err)
	}

	if err := a.checkAllFound(ctx, userID, categoryIDs, foundIDs); err != nil {
		return err
	}

	a.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "ValidateCategories"),
		observability.String("layer", "adapter"),
		observability.String("entity", "category"),
		observability.String("user_id", userID),
	)
	a.fm.RecordRepositoryQuery(ctx, "validate_categories", "category", time.Since(start))
	return nil
}

func (a *categoryProviderAdapter) queryFoundIDs(ctx context.Context, userID string, categoryIDs []string) (map[string]struct{}, error) {
	placeholders := make([]string, len(categoryIDs))
	args := make([]any, len(categoryIDs)+1)
	for i, id := range categoryIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	args[len(categoryIDs)] = userID

	query := fmt.Sprintf(
		"SELECT id FROM categories WHERE id IN (%s) AND user_id = $%d AND deleted_at IS NULL",
		strings.Join(placeholders, ", "),
		len(categoryIDs)+1,
	)

	rows, err := a.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	foundIDs := make(map[string]struct{}, len(categoryIDs))
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		foundIDs[id] = struct{}{}
	}

	return foundIDs, rows.Err()
}

func (a *categoryProviderAdapter) checkAllFound(ctx context.Context, userID string, categoryIDs []string, foundIDs map[string]struct{}) error {
	for _, id := range categoryIDs {
		if _, ok := foundIDs[id]; !ok {
			a.o11y.Logger().Warn(ctx, "category_not_found",
				observability.String("operation", "ValidateCategories"),
				observability.String("layer", "adapter"),
				observability.String("entity", "category"),
				observability.String("user_id", userID),
			)
			return pkginterfaces.ErrCategoryNotFound
		}
	}
	return nil
}

func (a *categoryProviderAdapter) logQueryFailed(ctx context.Context, userID string, err error) {
	a.o11y.Logger().Error(ctx, "query_failed",
		observability.String("operation", "ValidateCategories"),
		observability.String("layer", "adapter"),
		observability.String("entity", "category"),
		observability.String("user_id", userID),
		observability.Error(err),
	)
}
