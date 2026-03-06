package repositories

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financial/internal/card/domain/entities"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

func (r *cardRepository) Save(ctx context.Context, card *entities.Card) error {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "card_repository.save")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "save"),
		observability.String("layer", "repository"),
		observability.String("entity", "card"),
		observability.String("user_id", card.UserID.String()),
	)

	query := `insert into
				cards (
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
				)
				values
					($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "save"),
			observability.String("layer", "repository"),
			observability.String("entity", "card"),
			observability.String("user_id", card.UserID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "save", "card", "infra", time.Since(start))
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			span.RecordError(closeErr)
		}
	}()

	var dueDay any
	var closingOffset any
	if card.Type.IsCredit() {
		dueDay = card.DueDay.Value
		closingOffset = card.ClosingOffsetDays.Value
	}

	_, err = stmt.ExecContext(
		ctx,
		card.ID.Value,
		card.UserID.Value,
		card.Name.Value,
		card.Type.Value,
		card.Flag.Value,
		card.LastFourDigits.Value,
		dueDay,
		closingOffset,
		card.CreatedAt.Ptr(),
		card.UpdatedAt.Ptr(),
		card.DeletedAt.Ptr(),
	)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "save"),
			observability.String("layer", "repository"),
			observability.String("entity", "card"),
			observability.String("user_id", card.UserID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "save", "card", "infra", time.Since(start))
		return err
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "save"),
		observability.String("layer", "repository"),
		observability.String("entity", "card"),
		observability.String("user_id", card.UserID.String()),
	)
	r.fm.RecordRepositoryQuery(ctx, "save", "card", time.Since(start))
	return nil
}

func (r *cardRepository) Update(ctx context.Context, card *entities.Card) error {
	start := time.Now()
	ctx, span := r.o11y.Tracer().Start(ctx, "card_repository.update")
	defer span.End()

	r.o11y.Logger().Debug(ctx, "query_started",
		observability.String("operation", "update"),
		observability.String("layer", "repository"),
		observability.String("entity", "card"),
		observability.String("user_id", card.UserID.String()),
	)

	query := `update
				cards
			set
				name = $1,
				flag = $2,
				last_four_digits = $3,
				due_day = $4,
				closing_offset_days = $5,
				updated_at = $6,
				deleted_at = $7
			where
				id = $8
				and user_id = $9`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "update"),
			observability.String("layer", "repository"),
			observability.String("entity", "card"),
			observability.String("user_id", card.UserID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "update", "card", "infra", time.Since(start))
		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			span.RecordError(closeErr)
		}
	}()

	var dueDay any
	var closingOffset any
	if card.Type.IsCredit() {
		dueDay = card.DueDay.Value
		closingOffset = card.ClosingOffsetDays.Value
	}

	_, err = stmt.ExecContext(
		ctx,
		card.Name.Value,
		card.Flag.Value,
		card.LastFourDigits.Value,
		dueDay,
		closingOffset,
		card.UpdatedAt.Ptr(),
		card.DeletedAt.Ptr(),
		card.ID.Value,
		card.UserID.Value,
	)
	if err != nil {
		span.RecordError(err)
		r.o11y.Logger().Error(ctx, "query_failed",
			observability.String("operation", "update"),
			observability.String("layer", "repository"),
			observability.String("entity", "card"),
			observability.String("user_id", card.UserID.String()),
			observability.Error(err),
		)
		r.fm.RecordRepositoryFailure(ctx, "update", "card", "infra", time.Since(start))
		return err
	}

	r.o11y.Logger().Debug(ctx, "query_completed",
		observability.String("operation", "update"),
		observability.String("layer", "repository"),
		observability.String("entity", "card"),
		observability.String("user_id", card.UserID.String()),
	)
	r.fm.RecordRepositoryQuery(ctx, "update", "card", time.Since(start))
	return nil
}
