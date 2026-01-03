package messaging

import (
	"context"
)

// Handler processa mensagens de forma agnóstica ao broker.
//
// Handlers de domínio implementam esta interface para processar
// mensagens específicas. O lifecycle do handler é gerenciado pelo Consumer.
//
// Exemplo de uso:
//
//	type MyHandler struct {
//	    useCase MyUseCase
//	}
//
//	func (h *MyHandler) Handle(ctx context.Context, msg *Message) error {
//	    // Processar mensagem
//	    return h.useCase.Execute(ctx, msg.Payload)
//	}
//
//	func (h *MyHandler) Topics() []string {
//	    return []string{"order.created", "order.updated"}
//	}
type Handler interface {
	// Handle processa uma mensagem.
	// Retorna error para NACK/requeue, nil para ACK.
	// O contexto pode conter deadline/timeout configurado pelo consumer.
	Handle(ctx context.Context, msg *Message) error

	// Topics retorna lista de topics que este handler processa.
	// Pode incluir patterns dependendo do broker (ex: "order.*" para RabbitMQ).
	Topics() []string
}

// HandlerFunc é um adapter que permite usar funções como Handler.
// Útil para handlers simples que não precisam de struct.
//
// Exemplo de uso:
//
//	handler := messaging.NewFuncHandler(
//	    []string{"user.created"},
//	    func(ctx context.Context, msg *Message) error {
//	        // Processar mensagem
//	        return nil
//	    },
//	)
type HandlerFunc func(ctx context.Context, msg *Message) error

// funcHandler wrapper interno para HandlerFunc.
type funcHandler struct {
	fn     HandlerFunc
	topics []string
}

// NewFuncHandler cria um Handler a partir de uma função.
// topics: lista de topics que a função processa
// fn: função que processa a mensagem
func NewFuncHandler(topics []string, fn HandlerFunc) Handler {
	return &funcHandler{
		fn:     fn,
		topics: topics,
	}
}

// Handle implementa Handler interface.
func (h *funcHandler) Handle(ctx context.Context, msg *Message) error {
	return h.fn(ctx, msg)
}

// Topics implementa Handler interface.
func (h *funcHandler) Topics() []string {
	return h.topics
}
