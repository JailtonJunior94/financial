package repositories

import (
	"context"

	"github.com/jailtonjunior94/financial/internal/card/domain/entities"
	"github.com/jailtonjunior94/financial/internal/card/domain/interfaces"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type cardRepository struct {
	db   database.DBTX
	o11y observability.Observability
}

func NewCardRepository(db database.DBTX, o11y observability.Observability) interfaces.CardRepository {
	return &cardRepository{
		db:   db,
		o11y: o11y,
	}
}

func (r *cardRepository) List(ctx context.Context, userID vos.UUID) ([]*entities.Card, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "card_repository.list")
	defer span.End()

	query := `select
				id,
				user_id,
				name,
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
		err := rows.Scan(
			&card.ID.Value,
			&card.UserID.Value,
			&card.Name.Value,
			&card.DueDay.Value,
			&card.ClosingOffsetDays.Value,
			&card.CreatedAt,
			&card.UpdatedAt,
			&card.DeletedAt,
		)
		if err != nil {
			span.RecordError(err)

			return nil, err
		}
		cards = append(cards, &card)
	}
	return cards, nil
}

func (r *cardRepository) FindByID(ctx context.Context, userID, id vos.UUID) (*entities.Card, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "card_repository.find_by_id")
	defer span.End()

	query := `select
				id,
				user_id,
				name,
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
	err := r.db.QueryRowContext(ctx, query, userID.String(), id.String()).Scan(
		&card.ID.Value,
		&card.UserID.Value,
		&card.Name.Value,
		&card.DueDay.Value,
		&card.ClosingOffsetDays.Value,
		&card.CreatedAt,
		&card.UpdatedAt,
		&card.DeletedAt,
	)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		span.RecordError(err)

		return nil, err
	}

	return &card, nil
}

func (r *cardRepository) Save(ctx context.Context, card *entities.Card) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "card_repository.save")
	defer span.End()

	query := `insert into
				cards (
					id,
					user_id,
					name,
					due_day,
					closing_offset_days,
					created_at,
					updated_at,
					deleted_at
				)
				values
					($1, $2, $3, $4, $5, $6, $7, $8)`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		span.AddEvent(
			"error preparing insert card",
			observability.Field{Key: "user_id", Value: card.UserID},
			observability.Field{Key: "error", Value: err},
		)

		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			span.RecordError(closeErr)
		}
	}()

	_, err = stmt.ExecContext(
		ctx,
		card.ID.Value,
		card.UserID.Value,
		card.Name.Value,
		card.DueDay.Value,
		card.ClosingOffsetDays.Value,
		card.CreatedAt.Ptr(),
		card.UpdatedAt.Ptr(),
		card.DeletedAt.Ptr(),
	)
	if err != nil {
		span.AddEvent(
			"error inserting card",
			observability.Field{Key: "user_id", Value: card.UserID},
			observability.Field{Key: "error", Value: err},
		)

		return err
	}
	return nil
}

func (r *cardRepository) Update(ctx context.Context, card *entities.Card) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "card_repository.update")
	defer span.End()

	query := `update
				cards
			set
				name = $1,
				due_day = $2,
				closing_offset_days = $3,
				updated_at = $4,
				deleted_at = $5
			where
				id = $6
				and user_id = $7`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		span.AddEvent(
			"error preparing update card",
			observability.Field{Key: "user_id", Value: card.UserID},
			observability.Field{Key: "error", Value: err},
		)

		return err
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			span.RecordError(closeErr)
		}
	}()

	_, err = stmt.ExecContext(
		ctx,
		card.Name.Value,
		card.DueDay.Value,
		card.ClosingOffsetDays.Value,
		card.UpdatedAt.Ptr(),
		card.DeletedAt.Ptr(),
		card.ID.Value,
		card.UserID.Value,
	)
	if err != nil {
		span.AddEvent(
			"error updating card",
			observability.Field{Key: "user_id", Value: card.UserID},
			observability.Field{Key: "error", Value: err},
		)

		return err
	}

	return nil
}
