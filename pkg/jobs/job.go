package jobs

import (
	"context"
	"time"
)

// Job representa um cron job executável.
// Cada job deve implementar esta interface para ser agendado pelo scheduler.
type Job interface {
	// Run executa a lógica do job.
	// O contexto pode ser cancelado durante graceful shutdown.
	// Retorna error se o job falhar.
	Run(ctx context.Context) error

	// Name retorna o nome identificador do job (usado em logs e métricas).
	Name() string

	// Schedule retorna a expressão cron para agendamento.
	// Formato cron: "* * * * *" (minuto hora dia mês dia-da-semana).
	// Suporta também expressões especiais: @hourly, @daily, @weekly, @monthly.
	Schedule() string
}

// Config armazena configurações globais para execução de jobs.
type Config struct {
	// DefaultTimeout é o timeout padrão para execução de jobs.
	// Jobs que excederem este tempo serão cancelados.
	DefaultTimeout time.Duration

	// EnableRecovery habilita recovery automático para evitar crash em panic.
	EnableRecovery bool

	// MaxConcurrentJobs limita quantos jobs podem executar simultaneamente.
	// 0 = sem limite (não recomendado para produção).
	MaxConcurrentJobs int
}

// DefaultConfig retorna configuração padrão para jobs.
func DefaultConfig() *Config {
	return &Config{
		DefaultTimeout:    5 * time.Minute,
		EnableRecovery:    true,
		MaxConcurrentJobs: 10,
	}
}
