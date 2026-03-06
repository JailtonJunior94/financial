package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jailtonjunior94/financial/internal/card/domain/entities"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

func (r *cardRepository) FindByIDOnly(ctx context.Context, id vos.UUID) (*entities.Card, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "card_repository.find_by_id_only")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "find_by_id_only"),
		observability.String("layer", "repository"),
		observability.String("entity", "card"),
		observability.String("card_id", id.String()),
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
				deleted_at is null
				and id = $1;`

	var card entities.Card
	var dueDayNull sql.NullInt32
	var closingOffsetNull sql.NullInt32

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
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
		if errors.Is(err, sql.ErrNoRows) {
			r.o11y.Logger().Debug(ctx, "query_completed",
				observability.String("operation", "find_by_id_only"),
				observability.String("layer", "repository"),
				observability.String("entity", "card"),
				observability.String("card_id", id.String()),
			)
			r.fm.RecordRepositoryQuery(ctx, "find_by_id_only", "card", time.Since(start))
			return nil, nil
		}
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "find_by_id_only"),
			observability.String("layer", "repository"),
			observability.String("entity", "card"),
			observability.String("card_id", id.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "find_by_id_only", "card", "infra", time.Since(start))
		return nil, err
	}

	if dueDayNull.Valid {
		card.DueDay.Value = int(dueDayNull.Int32)
	}
	if closingOffsetNull.Valid {
		card.ClosingOffsetDays.Value = int(closingOffsetNull.Int32)
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "find_by_id_only"),
		observability.String("layer", "repository"),
		observability.String("entity", "card"),
		observability.String("card_id", id.String()),
	)
	r.fm.RecordRepositoryQuery(ctx, "find_by_id_only", "card", time.Since(start))
	return &card, nil
}

func (r *cardRepository) FindByID(ctx context.Context, userID, id vos.UUID) (*entities.Card, error) {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "card_repository.find_by_id")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "find_by_id"),
		observability.String("layer", "repository"),
		observability.String("entity", "card"),
		observability.String("user_id", userID.String()),
		observability.String("card_id", id.String()),
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
				and id = $2;`

	var card entities.Card
	var dueDayNull sql.NullInt32
	var closingOffsetNull sql.NullInt32

	err := r.db.QueryRowContext(ctx, query, userID.String(), id.String()).Scan(
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
		if errors.Is(err, sql.ErrNoRows) {
			r.o11y.Logger().Debug(ctx, "query_completed",
				observability.String("operation", "find_by_id"),
				observability.String("layer", "repository"),
				observability.String("entity", "card"),
				observability.String("user_id", userID.String()),
				observability.String("card_id", id.String()),
			)
			r.fm.RecordRepositoryQuery(ctx, "find_by_id", "card", time.Since(start))
			return nil, nil
		}
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "find_by_id"),
			observability.String("layer", "repository"),
			observability.String("entity", "card"),
			observability.String("user_id", userID.String()),
			observability.String("card_id", id.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "find_by_id", "card", "infra", time.Since(start))
		return nil, err
	}

	if dueDayNull.Valid {
		card.DueDay.Value = int(dueDayNull.Int32)
	}
	if closingOffsetNull.Valid {
		card.ClosingOffsetDays.Value = int(closingOffsetNull.Int32)
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "find_by_id"),
		observability.String("layer", "repository"),
		observability.String("entity", "card"),
		observability.String("user_id", userID.String()),
		observability.String("card_id", id.String()),
	)
	r.fm.RecordRepositoryQuery(ctx, "find_by_id", "card", time.Since(start))
	return &card, nil
}
