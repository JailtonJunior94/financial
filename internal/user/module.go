package user

import (
	"github.com/jailtonjunior94/financial/internal/user/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/user/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/internal/user/usecase"
	"github.com/jailtonjunior94/financial/pkg/bundle"

	"github.com/go-chi/chi/v5"
)

func RegisterAuthModule(ioc *bundle.Container, router *chi.Mux) {
	userRepository := repositories.NewUserRepository(ioc.DB, ioc.Observability)
	authUseCase := usecase.NewTokenUseCase(ioc.Config, ioc.Observability, ioc.Hash, ioc.Jwt, userRepository)
	authHandler := http.NewAuthHandler(ioc.Observability, authUseCase)
	http.NewAuthRoute(router, http.WithTokenHandler(authHandler.Token))
}

func RegisterUserModule(ioc *bundle.Container, router *chi.Mux) {
	userRepository := repositories.NewUserRepository(ioc.DB, ioc.Observability)
	createUserUseCase := usecase.NewCreateUserUseCase(ioc.Observability, ioc.Hash, userRepository)
	userHandler := http.NewUserHandler(ioc.Observability, createUserUseCase)
	http.NewUserRoute(
		router,
		http.WithCreateUserHandler(userHandler.Create),
	)
}
