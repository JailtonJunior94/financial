package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/pkg/jobs"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

// DatabaseCleanupJob é um job exemplo que faz limpeza de registros antigos no banco de dados.
// Demonstra uso de conexão DB em cron job com tratamento de erros e logging estruturado.
type DatabaseCleanupJob struct {
	db   *sql.DB
	o11y observability.Observability
}

// NewDatabaseCleanupJob cria uma nova instância do job de limpeza.
func NewDatabaseCleanupJob(db *sql.DB, o11y observability.Observability) jobs.Job {
	return &DatabaseCleanupJob{
		db:   db,
		o11y: o11y,
	}
}

// Name retorna o identificador do job.
func (j *DatabaseCleanupJob) Name() string {
	return "database_cleanup"
}

// Schedule retorna a expressão cron.
// Executa diariamente às 2h da manhã (horário ideal para manutenção).
func (j *DatabaseCleanupJob) Schedule() string {
	return "0 2 * * *" // 02:00 todos os dias
}

// Run executa a lógica de limpeza do banco de dados.
// Remove registros soft-deleted há mais de 90 dias.
func (j *DatabaseCleanupJob) Run(ctx context.Context) error {
	j.o11y.Logger().Info(ctx, "starting database cleanup")

	// Define período de retenção (90 dias)
	retentionPeriod := 90 * 24 * time.Hour
	cutoffDate := time.Now().Add(-retentionPeriod)

	// Exemplo: Limpar categorias deletadas há mais de 90 dias
	// IMPORTANTE: Adapte as queries de acordo com suas tabelas
	deletedCount, err := j.cleanupCategories(ctx, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to cleanup categories: %w", err)
	}

	j.o11y.Logger().Info(ctx, "database cleanup completed",
		observability.Int("categories_deleted", deletedCount),
		observability.String("cutoff_date", cutoffDate.Format(time.RFC3339)),
	)

	return nil
}

// cleanupCategories remove permanentemente categorias soft-deleted antigas.
func (j *DatabaseCleanupJob) cleanupCategories(ctx context.Context, cutoffDate time.Time) (int, error) {
	query := `
		DELETE FROM categories 
		WHERE deleted_at IS NOT NULL 
		  AND deleted_at < $1
	`

	result, err := j.db.ExecContext(ctx, query, cutoffDate)
	if err != nil {
		return 0, fmt.Errorf("exec cleanup query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}
