package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jailtonjunior94/financial/pkg/jobs"

	"github.com/JailtonJunior94/devkit-go/pkg/messaging/rabbitmq"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

// ReportGeneratorJob é um job exemplo que gera relatórios e publica no RabbitMQ.
// Demonstra uso de DB para consultar dados e RabbitMQ para publicar eventos.
type ReportGeneratorJob struct {
	db        *sql.DB
	publisher *rabbitmq.Publisher
	exchange  string
	o11y      observability.Observability
}

// ReportData representa os dados de um relatório.
type ReportData struct {
	ReportID      string    `json:"report_id"`
	GeneratedAt   time.Time `json:"generated_at"`
	TotalUsers    int       `json:"total_users"`
	ActiveBudgets int       `json:"active_budgets"`
	TotalInvoices int       `json:"total_invoices"`
}

// NewReportGeneratorJob cria uma nova instância do job de geração de relatórios.
func NewReportGeneratorJob(db *sql.DB, rabbit *rabbitmq.Client, exchange string, o11y observability.Observability) jobs.Job {
	return &ReportGeneratorJob{
		db:        db,
		publisher: rabbitmq.NewPublisher(rabbit),
		exchange:  exchange,
		o11y:      o11y,
	}
}

// Name retorna o identificador do job.
func (j *ReportGeneratorJob) Name() string {
	return "report_generator"
}

// Schedule retorna a expressão cron.
// Executa toda segunda-feira às 8h (início da semana).
func (j *ReportGeneratorJob) Schedule() string {
	return "0 8 * * 1" // 08:00 todas as segundas-feiras
}

// Run executa a lógica de geração de relatórios.
// Coleta estatísticas do banco e publica evento no RabbitMQ.
func (j *ReportGeneratorJob) Run(ctx context.Context) error {
	j.o11y.Logger().Info(ctx, "starting report generation")

	// 1. Coletar dados do banco de dados
	report, err := j.collectReportData(ctx)
	if err != nil {
		return fmt.Errorf("failed to collect report data: %w", err)
	}

	j.o11y.Logger().Info(ctx, "report data collected",
		observability.Int("total_users", report.TotalUsers),
		observability.Int("active_budgets", report.ActiveBudgets),
		observability.Int("total_invoices", report.TotalInvoices),
	)

	// 2. Serializar dados para JSON
	payload, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	// 3. Publicar evento no RabbitMQ
	routingKey := "reports.weekly.generated"
	if err := j.publisher.Publish(ctx, j.exchange, routingKey, payload,
		rabbitmq.WithContentType("application/json"),
		rabbitmq.WithDeliveryMode(2), // Persistent
	); err != nil {
		return fmt.Errorf("failed to publish report: %w", err)
	}

	j.o11y.Logger().Info(ctx, "report published to rabbitmq",
		observability.String("exchange", j.exchange),
		observability.String("routing_key", routingKey),
		observability.String("report_id", report.ReportID),
	)

	return nil
}

// collectReportData coleta estatísticas do banco de dados.
func (j *ReportGeneratorJob) collectReportData(ctx context.Context) (*ReportData, error) {
	report := &ReportData{
		ReportID:    fmt.Sprintf("REPORT-%d", time.Now().Unix()),
		GeneratedAt: time.Now(),
	}

	// Contar total de usuários
	if err := j.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&report.TotalUsers); err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	// Contar budgets ativos
	if err := j.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM budgets WHERE deleted_at IS NULL").Scan(&report.ActiveBudgets); err != nil {
		return nil, fmt.Errorf("count budgets: %w", err)
	}

	// Contar faturas totais
	if err := j.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM invoices WHERE deleted_at IS NULL").Scan(&report.TotalInvoices); err != nil {
		return nil, fmt.Errorf("count invoices: %w", err)
	}

	return report, nil
}
