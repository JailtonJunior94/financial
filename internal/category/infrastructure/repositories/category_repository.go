package repositories

import (
"context"
"database/sql"
"fmt"
"time"

"github.com/jailtonjunior94/financial/internal/category/domain/entities"
"github.com/jailtonjunior94/financial/internal/category/domain/interfaces"
"github.com/jailtonjunior94/financial/pkg/observability/metrics"

"github.com/JailtonJunior94/devkit-go/pkg/database"
"github.com/JailtonJunior94/devkit-go/pkg/observability"
"github.com/JailtonJunior94/devkit-go/pkg/vos"
)

type categoryRepository struct {
db   database.DBTX
o11y observability.Observability
fm   *metrics.FinancialMetrics
}

func NewCategoryRepository(db database.DBTX, o11y observability.Observability, fm *metrics.FinancialMetrics) interfaces.CategoryRepository {
return &categoryRepository{
db:   db,
o11y: o11y,
fm:   fm,
}
}

func (r *categoryRepository) ListPaginated(ctx context.Context, params interfaces.ListCategoriesParams) ([]*entities.Category, error) {
start := time.Now()
ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.list_paginated")
defer span.End()

whereClause := "user_id = $1 AND deleted_at IS NULL"
args := []interface{}{params.UserID.String()}

cursorSequence, hasSeq := params.Cursor.GetInt("sequence")
cursorID, hasID := params.Cursor.GetString("id")

if hasSeq && hasID && cursorID != "" {
whereClause += ` AND (sequence > $2 OR (sequence = $2 AND id > $3))`
args = append(args, cursorSequence, cursorID)
}

query := fmt.Sprintf(`
SELECT id, user_id, name, sequence, created_at, updated_at, deleted_at
FROM categories
WHERE %s
ORDER BY sequence ASC, id ASC
LIMIT $%d`, whereClause, len(args)+1)

args = append(args, params.Limit)

rows, err := r.db.QueryContext(ctx, query, args...)
if err != nil {
span.RecordError(err)
r.fm.RecordRepositoryFailure(ctx, "list_paginated", "category", "infra", time.Since(start))
return nil, err
}
defer func() { _ = rows.Close() }()

categories := make([]*entities.Category, 0)
for rows.Next() {
var category entities.Category
if err := rows.Scan(
&category.ID.Value,
&category.UserID.Value,
&category.Name.Value,
&category.Sequence.Sequence,
&category.CreatedAt,
&category.UpdatedAt,
&category.DeletedAt,
); err != nil {
span.RecordError(err)
r.fm.RecordRepositoryFailure(ctx, "list_paginated", "category", "infra", time.Since(start))
return nil, err
}
categories = append(categories, &category)
}

if err := rows.Err(); err != nil {
span.RecordError(err)
r.fm.RecordRepositoryFailure(ctx, "list_paginated", "category", "infra", time.Since(start))
return nil, err
}

r.fm.RecordRepositoryQuery(ctx, "list_paginated", "category", time.Since(start))
return categories, nil
}

func (r *categoryRepository) FindByID(ctx context.Context, userID, id vos.UUID) (*entities.Category, error) {
start := time.Now()
ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.find_by_id")
defer span.End()

query := `
SELECT id, user_id, name, sequence, created_at, updated_at, deleted_at
FROM categories
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`

var category entities.Category
err := r.db.QueryRowContext(ctx, query, id.String(), userID.String()).Scan(
&category.ID.Value,
&category.UserID.Value,
&category.Name.Value,
&category.Sequence.Sequence,
&category.CreatedAt,
&category.UpdatedAt,
&category.DeletedAt,
)

if err == sql.ErrNoRows {
r.fm.RecordRepositoryQuery(ctx, "find_by_id", "category", time.Since(start))
return nil, nil
}
if err != nil {
span.RecordError(err)
r.fm.RecordRepositoryFailure(ctx, "find_by_id", "category", "infra", time.Since(start))
return nil, fmt.Errorf("category_repository.find_by_id: %w", err)
}

r.fm.RecordRepositoryQuery(ctx, "find_by_id", "category", time.Since(start))
return &category, nil
}

func (r *categoryRepository) Save(ctx context.Context, category *entities.Category) error {
start := time.Now()
ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.save")
defer span.End()

query := `INSERT INTO categories (id, user_id, name, sequence, created_at)
          VALUES ($1, $2, $3, $4, $5)`

_, err := r.db.ExecContext(ctx, query,
category.ID.Value,
category.UserID.Value,
category.Name.Value,
category.Sequence.Sequence,
category.CreatedAt.Ptr(),
)
if err != nil {
span.RecordError(err)
r.fm.RecordRepositoryFailure(ctx, "save", "category", "infra", time.Since(start))
return fmt.Errorf("category_repository.save: %w", err)
}

r.fm.RecordRepositoryQuery(ctx, "save", "category", time.Since(start))
return nil
}

func (r *categoryRepository) Update(ctx context.Context, category *entities.Category) error {
start := time.Now()
ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.update")
defer span.End()

query := `UPDATE categories SET name = $1, sequence = $2, updated_at = $3
          WHERE id = $4 AND deleted_at IS NULL`

_, err := r.db.ExecContext(ctx, query,
category.Name.Value,
category.Sequence.Sequence,
category.UpdatedAt.Ptr(),
category.ID.Value,
)
if err != nil {
span.RecordError(err)
r.fm.RecordRepositoryFailure(ctx, "update", "category", "infra", time.Since(start))
return fmt.Errorf("category_repository.update: %w", err)
}

r.fm.RecordRepositoryQuery(ctx, "update", "category", time.Since(start))
return nil
}

func (r *categoryRepository) SoftDelete(ctx context.Context, id vos.UUID) error {
start := time.Now()
ctx, span := r.o11y.Tracer().Start(ctx, "category_repository.soft_delete")
defer span.End()

_, err := r.db.ExecContext(ctx,
`UPDATE categories SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
id,
)
if err != nil {
span.RecordError(err)
r.fm.RecordRepositoryFailure(ctx, "soft_delete", "category", "infra", time.Since(start))
return fmt.Errorf("category_repository.soft_delete: %w", err)
}

r.fm.RecordRepositoryQuery(ctx, "soft_delete", "category", time.Since(start))
return nil
}
