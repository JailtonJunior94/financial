package user

import (
	"github.com/jailtonjunior94/financial/internal/user/application/usecase"
	"github.com/jailtonjunior94/financial/internal/user/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/user/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/bundle"

	"github.com/JailtonJunior94/devkit-go/pkg/httpserver"
)

func RegisterAuthModule(ioc *bundle.Container) []httpserver.Route {
	userRepository := repositories.NewUserRepository(ioc.DB, ioc.Observability)
	authUseCase := usecase.NewTokenUseCase(ioc.Config, ioc.Observability, ioc.Hash, ioc.Jwt, userRepository)
	authHandler := http.NewAuthHandler(ioc.Observability, authUseCase)

	authRoutes := http.NewUserRoutes()
	authRoutes.Register(
		httpserver.NewRoute(
			"POST",
			"/api/v1/token",
			authHandler.Token,
		),
	)
	return authRoutes.Routes()
}

func RegisterUserModule(ioc *bundle.Container) []httpserver.Route {
	userRepository := repositories.NewUserRepository(ioc.DB, ioc.Observability)
	createUserUseCase := usecase.NewCreateUserUseCase(ioc.Observability, ioc.Hash, userRepository)
	userHandler := http.NewUserHandler(ioc.Observability, createUserUseCase)

	userRoutes := http.NewUserRoutes()
	userRoutes.Register(
		httpserver.NewRoute(
			"POST",
			"/api/v1/users",
			userHandler.Create,
		),
	)
	return userRoutes.Routes()
}
