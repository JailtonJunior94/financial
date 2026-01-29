package database

import (
	"context"
	"time"

	postgres "github.com/JailtonJunior94/devkit-go/pkg/database/postgres_otelsql"
)

// DatabaseOption é uma função que configura a conexão do banco de dados
type DatabaseOption func(*postgres.Config)

// WithDSN configura a connection string do banco de dados
func WithDSN(dsn string) DatabaseOption {
	return func(c *postgres.Config) {
		c.DSN = dsn
	}
}

// WithMaxOpenConns configura o número máximo de conexões abertas
func WithMaxOpenConns(n int) DatabaseOption {
	return func(c *postgres.Config) {
		c.MaxOpenConns = n
	}
}

// WithMaxIdleConns configura o número máximo de conexões ociosas
func WithMaxIdleConns(n int) DatabaseOption {
	return func(c *postgres.Config) {
		c.MaxIdleConns = n
	}
}

// WithConnMaxLifetime configura o tempo máximo de vida de uma conexão
func WithConnMaxLifetime(d time.Duration) DatabaseOption {
	return func(c *postgres.Config) {
		c.ConnMaxLifetime = d
	}
}

// WithConnMaxIdleTime configura o tempo máximo de ociosidade de uma conexão
func WithConnMaxIdleTime(d time.Duration) DatabaseOption {
	return func(c *postgres.Config) {
		c.ConnMaxIdleTime = d
	}
}

// WithMetrics habilita/desabilita métricas automáticas do pool
func WithMetrics(enabled bool) DatabaseOption {
	return func(c *postgres.Config) {
		c.EnableMetrics = enabled
	}
}

// WithQueryLogging habilita/desabilita SQL commenter (logs de queries)
func WithQueryLogging(enabled bool) DatabaseOption {
	return func(c *postgres.Config) {
		c.EnableQueryLogging = enabled
	}
}

func WithServiceName(serviceName string) DatabaseOption {
	return func(c *postgres.Config) {
		c.ServiceName = serviceName
	}
}

// NewDatabaseManager cria um novo gerenciador de banco de dados com options pattern
func NewDatabaseManager(ctx context.Context, opts ...DatabaseOption) (*postgres.DBManager, error) {
	config := &postgres.Config{
		ServiceName:        "financial-service",
		MaxOpenConns:       25,
		MaxIdleConns:       5,
		ConnMaxLifetime:    15 * time.Minute,
		ConnMaxIdleTime:    5 * time.Minute,
		EnableMetrics:      true,
		EnableQueryLogging: false,
	}

	// Aplicar options
	for _, opt := range opts {
		opt(config)
	}

	return postgres.NewDBManager(ctx, config)
}
