package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/observability"
)

// GroupConfig configuração do gerenciador de grupo.
type GroupConfig struct {
	// ShutdownTimeout timeout para shutdown de todos os services.
	ShutdownTimeout time.Duration

	// StartTimeout timeout para start de todos os services.
	StartTimeout time.Duration
}

// DefaultGroupConfig retorna configuração padrão do grupo.
func DefaultGroupConfig() *GroupConfig {
	return &GroupConfig{
		ShutdownTimeout: 30 * time.Second,
		StartTimeout:    10 * time.Second,
	}
}

// Group gerencia múltiplos services com lifecycle unificado.
// Permite start ordenado e shutdown graceful coordenado de vários services.
//
// Exemplo de uso:
//
//	group := lifecycle.NewGroup(o11y, lifecycle.DefaultGroupConfig())
//	group.Register(service1)
//	group.Register(service2)
//
//	if err := group.Start(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Wait for shutdown signal...
//	group.Shutdown(shutdownCtx)
type Group struct {
	services []Service
	config   *GroupConfig
	o11y     observability.Observability
	mu       sync.RWMutex
}

// NewGroup cria um novo gerenciador de grupo de services.
func NewGroup(o11y observability.Observability, config *GroupConfig) *Group {
	if config == nil {
		config = DefaultGroupConfig()
	}

	return &Group{
		services: make([]Service, 0),
		config:   config,
		o11y:     o11y,
	}
}

// Register adiciona um service ao grupo.
// Services serão iniciados na ordem de registro.
// Thread-safe: pode ser chamado concorrentemente.
func (g *Group) Register(service Service) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.services = append(g.services, service)

	g.o11y.Logger().Info(context.Background(),
		"service registered",
		observability.String("service", service.Name()),
		observability.Int("total_services", len(g.services)),
	)
}

// Start inicia todos os services registrados sequencialmente.
// Services são iniciados na ordem de registro.
// Se algum service falhar ao iniciar, retorna erro imediatamente.
func (g *Group) Start(ctx context.Context) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Cria contexto com timeout para start
	startCtx, cancel := context.WithTimeout(ctx, g.config.StartTimeout)
	defer cancel()

	g.o11y.Logger().Info(startCtx,
		"starting service group",
		observability.Int("services_count", len(g.services)),
	)

	// Start services sequencialmente
	for _, service := range g.services {
		serviceName := service.Name()

		g.o11y.Logger().Debug(startCtx,
			"starting service",
			observability.String("service", serviceName),
		)

		start := time.Now()

		if err := service.Start(startCtx); err != nil {
			g.o11y.Logger().Error(startCtx,
				"failed to start service",
				observability.String("service", serviceName),
				observability.Error(err),
			)
			return fmt.Errorf("failed to start service %s: %w", serviceName, err)
		}

		duration := time.Since(start)

		g.o11y.Logger().Info(startCtx,
			"service started",
			observability.String("service", serviceName),
			observability.Int64("duration_ms", duration.Milliseconds()),
		)
	}

	g.o11y.Logger().Info(startCtx,
		"service group started successfully",
		observability.Int("services_count", len(g.services)),
	)

	return nil
}

// Shutdown para todos os services gracefully.
// Services são parados em ordem reversa (LIFO) de forma paralela.
// Aguarda todos os services pararem ou timeout expirar.
func (g *Group) Shutdown(ctx context.Context) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Cria contexto com timeout para shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, g.config.ShutdownTimeout)
	defer cancel()

	g.o11y.Logger().Info(shutdownCtx,
		"shutting down service group",
		observability.Int("services_count", len(g.services)),
	)

	var wg sync.WaitGroup
	errors := make(chan error, len(g.services))

	// Shutdown em ordem reversa (LIFO) de forma paralela
	for i := len(g.services) - 1; i >= 0; i-- {
		service := g.services[i]
		wg.Add(1)

		go func(svc Service) {
			defer wg.Done()

			serviceName := svc.Name()

			g.o11y.Logger().Debug(shutdownCtx,
				"shutting down service",
				observability.String("service", serviceName),
			)

			start := time.Now()

			if err := svc.Shutdown(shutdownCtx); err != nil {
				g.o11y.Logger().Error(shutdownCtx,
					"service shutdown error",
					observability.String("service", serviceName),
					observability.Error(err),
				)
				errors <- fmt.Errorf("service %s shutdown error: %w", serviceName, err)
				return
			}

			duration := time.Since(start)

			g.o11y.Logger().Info(shutdownCtx,
				"service stopped",
				observability.String("service", serviceName),
				observability.Int64("duration_ms", duration.Milliseconds()),
			)
		}(service)
	}

	// Aguarda shutdown completo
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
		close(errors)
	}()

	select {
	case <-done:
		// Coleta erros
		var shutdownErrors []error
		for err := range errors {
			shutdownErrors = append(shutdownErrors, err)
		}

		if len(shutdownErrors) > 0 {
			g.o11y.Logger().Error(shutdownCtx,
				"shutdown completed with errors",
				observability.Int("error_count", len(shutdownErrors)),
			)
			return fmt.Errorf("shutdown errors: %v", shutdownErrors)
		}

		g.o11y.Logger().Info(shutdownCtx, "service group shutdown completed successfully")
		return nil

	case <-shutdownCtx.Done():
		g.o11y.Logger().Warn(shutdownCtx,
			"shutdown timeout exceeded",
			observability.Int("services_count", len(g.services)),
		)
		return fmt.Errorf("shutdown timeout exceeded")
	}
}
