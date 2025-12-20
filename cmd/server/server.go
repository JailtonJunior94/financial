package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jailtonjunior94/financial/internal/category"
	"github.com/jailtonjunior94/financial/internal/user"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/bundle"
	customErrors "github.com/jailtonjunior94/financial/pkg/custom_errors"

	"github.com/JailtonJunior94/devkit-go/pkg/httpserver"
	"github.com/JailtonJunior94/devkit-go/pkg/responses"
)

func Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ioc := bundle.NewContainer(ctx)

	/* Observability */
	defer func() {
		if err := ioc.Telemetry.Shutdown(ctx); err != nil {
			log.Printf("erro ao finalizar a telemetria: %v", err)
		}
	}()

	/* Close DBConnection */
	defer func() {
		if err := ioc.DB.Close(); err != nil {
			log.Printf("erro ao fechar conexão com banco de dados: %v", err)
		}
	}()

	healthRoute := httpserver.NewRoute(http.MethodGet, "/health", func(w http.ResponseWriter, r *http.Request) error {
		if err := ioc.DB.Ping(); err != nil {
			return err
		}
		responses.JSON(w, http.StatusOK, map[string]any{"status": "ok"})
		return nil
	})

	routes := []httpserver.Route{healthRoute}
	authRoutes := user.RegisterAuthModule(ioc)
	userRoutes := user.RegisterUserModule(ioc)
	categoryRoutes := category.RegisterCategoryModule(ioc)

	routes = append(routes, authRoutes...)
	routes = append(routes, userRoutes...)
	routes = append(routes, categoryRoutes...)

	server := httpserver.New(
		httpserver.WithPort(ioc.Config.HTTPConfig.Port),
		httpserver.WithRoutes(routes...),
		httpserver.WithMiddlewares(
			httpserver.RequestID,
		),
		httpserver.WithErrorHandler(func(ctx context.Context, w http.ResponseWriter, err error) {
			// Se for CustomError, extrair o erro original
			if customErr, ok := err.(*customErrors.CustomError); ok {
				// Usar o erro original para buscar o código HTTP correto
				responseErr := httperrors.GetResponseError(customErr.Err)

				// Se tiver detalhes, incluir na resposta
				if customErr.Details != nil {
					responses.ErrorWithDetails(w, responseErr.Code, customErr.Message, customErr.Details)
					return
				}

				// Usar a mensagem do CustomError (mais específica) ou do mapping
				message := customErr.Message
				if message == "" {
					message = responseErr.Message
				}
				responses.Error(w, responseErr.Code, message)
				return
			}

			// Para erros não customizados, tentar mapear diretamente
			responseErr := httperrors.GetResponseError(err)
			responses.Error(w, responseErr.Code, responseErr.Message)
		}),
	)

	shutdown := server.Run()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := <-server.ShutdownListener(); err != nil && err != http.ErrServerClosed {
			interrupt <- syscall.SIGTERM
		}
	}()

	<-interrupt
	cancel() // Cancel context to propagate shutdown signal

	if err := shutdown(ctx); err != nil {
		log.Printf("erro ao finalizar o servidor: %v", err)
	}
}
