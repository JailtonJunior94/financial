package lifecycle

import "context"

// Service representa qualquer componente com lifecycle gerenciável.
// Componentes como Jobs, Consumers, Workers e outros serviços de background
// implementam esta interface para permitir gerenciamento uniforme de lifecycle.
//
// Exemplo de uso:
//
//	type MyService struct {}
//
//	func (s *MyService) Start(ctx context.Context) error {
//	    // Inicializar serviço
//	    return nil
//	}
//
//	func (s *MyService) Shutdown(ctx context.Context) error {
//	    // Parar serviço gracefully
//	    return nil
//	}
//
//	func (s *MyService) Name() string {
//	    return "my-service"
//	}
type Service interface {
	// Start inicia o serviço.
	// Deve retornar imediatamente (não-bloqueante).
	// Use goroutines para operações de longa duração.
	Start(ctx context.Context) error

	// Shutdown para o serviço gracefully.
	// Deve respeitar o timeout definido no contexto.
	// Retorna error se não conseguir parar dentro do timeout.
	Shutdown(ctx context.Context) error

	// Name retorna identificador único do serviço (usado em logs).
	Name() string
}
