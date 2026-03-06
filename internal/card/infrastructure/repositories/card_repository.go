package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/internal/card/domain/entities"
	"github.com/jailtonjunior94/financial/internal/card/domain/interfaces"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type cardRepository struct {
	db   database.DBTX
	o11y observability.Observability
	fm   *metrics.FinancialMetrics
}

func NewCardRepository(db database.DBTX, o11y observability.Observability, fm *metrics.FinancialMetrics) interfaces.CardRepository {
	return &cardRepository{
		db:   db,
		o11y: o11y,
		fm:   fm,
	}
}

func (r *cardRepository) List(ctx context.Context, userID vos.UUID) ([]*entities.Card, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "card_repository.list")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "list"),
		observability.String("layer", "repository"),
		observability.String("entity", "card"),
		observability.String("user_id", userID.String()),
	)

	query := `select
				id,
				user_id,
				name,
				type,
				flag,
				last_four_digits,
				due_day,
				closing_offset_days,
				created_at,
				updated_at,
				deleted_at
			from
				cards
			where
				user_id = $1
				and deleted_at is null
			order by
				name;`

	rows, err := r.db.QueryContext(ctx, query, userID.String())
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "list"),
			observability.String("layer", "repository"),
			observability.String("entity", "card"),
			observability.String("user_id", userID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "list", "card", "infra", time.Since(start))
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			span.RecordError(closeErr)
		}
	}()

	var cards []*entities.Card
	for rows.Next() {
		var card entities.Card
		var dueDayNull sql.NullInt32
		var closingOffsetNull sql.NullInt32

		err := rows.Scan(
			&card.ID.Value,
			&card.UserID.Value,
			&card.Name.Value,
			&card.Type.Value,
			&card.Flag.Value,
			&card.LastFourDigits.Value,
			&dueDayNull,
			&closingOffsetNull,
			&card.CreatedAt,
			&card.UpdatedAt,
			&card.DeletedAt,
		)
		if err != nil {
			span.RecordError(err)
			r.o11y.Logger().Error(ctx, "query_failed",
				observability.String("operation", "list"),
				observability.String("layer", "repository"),
				observability.String("entity", "card"),
				observability.String("user_id", userID.String()),
				observability.Error(err),
			)
			r.fm.RecordRepositoryFailure(ctx, "list", "card", "infra", time.Since(start))
			return nil, err
		}
		if dueDayNull.Valid {
			card.DueDay.Value = int(dueDayNull.Int32)
		}
		if closingOffsetNull.Valid {
			card.ClosingOffsetDays.Value = int(closingOffsetNull.Int32)
		}
		cards = append(cards, &card)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "list"),
			observability.String("layer", "repository"),
			observability.String("entity", "card"),
			observability.String("user_id", userID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "list", "card", "infra", time.Since(start))
		return nil, err
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "list"),
		observability.String("layer", "repository"),
		observability.String("entity", "card"),
		observability.String("user_id", userID.String()),
	)
	r.fm.RecordRepositoryQuery(ctx, "list", "card", time.Since(start))
	return cards, nil
}

func (r *cardRepository) ListPaginated(ctx context.Context, params interfaces.ListCardsParams) ([]*entities.Card, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "card_repository.list_paginated")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "list_paginated"),
		observability.String("layer", "repository"),
		observability.String("entity", "card"),
		observability.String("user_id", params.UserID.String()),
	)

	whereClause := "user_id = $1 AND deleted_at IS NULL"
	args := []any{params.UserID.String()}

	cursorName, hasName := params.Cursor.GetString("name")
	cursorID, hasID := params.Cursor.GetString("id")

	if hasName && hasID && cursorName != "" && cursorID != "" {
		whereClause += ` AND (
			name > $2
			OR (name = $2 AND id > $3)
		)`
		args = append(args, cursorName, cursorID)
	}

	query := fmt.Sprintf(`
		SELECT
			id,
			user_id,
			name,
			type,
			flag,
			last_four_digits,
			due_day,
			closing_offset_days,
			created_at,
			updated_at,
			deleted_at
		FROM cards
		WHERE %s
		ORDER BY name ASC, id ASC
		LIMIT $%d`, whereClause, len(args)+1)

	args = append(args, params.Limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "list_paginated"),
			observability.String("layer", "repository"),
			observability.String("entity", "card"),
			observability.String("user_id", params.UserID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "list_paginated", "card", "infra", time.Since(start))
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			span.RecordError(closeErr)
		}
	}()

	cards := make([]*entities.Card, 0)
	for rows.Next() {
		var card entities.Card
		var dueDayNull sql.NullInt32
		var closingOffsetNull sql.NullInt32

		err := rows.Scan(
			&card.ID.Value,
			&card.UserID.Value,
			&card.Name.Value,
			&card.Type.Value,
			&card.Flag.Value,
			&card.LastFourDigits.Value,
			&dueDayNull,
			&closingOffsetNull,
			&card.CreatedAt,
			&card.UpdatedAt,
			&card.DeletedAt,
		)
		if err != nil {
			span.RecordError(err)
			r.o11y.Logger().Error(ctx, "query_failed",
				observability.String("operation", "list_paginated"),
				observability.String("layer", "repository"),
				observability.String("entity", "card"),
				observability.String("user_id", params.UserID.String()),
				observability.Error(err),
			)
			r.fm.RecordRepositoryFailure(ctx, "list_paginated", "card", "infra", time.Since(start))
			return nil, err
		}
		if dueDayNull.Valid {
			card.DueDay.Value = int(dueDayNull.Int32)
		}
		if closingOffsetNull.Valid {
			card.ClosingOffsetDays.Value = int(closingOffsetNull.Int32)
		}
		cards = append(cards, &card)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "list_paginated"),
			observability.String("layer", "repository"),
			observability.String("entity", "card"),
			observability.String("user_id", params.UserID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "list_paginated", "card", "infra", time.Since(start))
		return nil, err
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "list_paginated"),
		observability.String("layer", "repository"),
		observability.String("entity", "card"),
		observability.String("user_id", params.UserID.String()),
	)
	r.fm.RecordRepositoryQuery(ctx, "list_paginated", "card", time.Since(start))
	return cards, nil
}
