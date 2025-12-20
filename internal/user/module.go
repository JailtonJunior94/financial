package user

import (
	"github.com/jailtonjunior94/financial/internal/user/application/usecase"
	"github.com/jailtonjunior94/financial/internal/user/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/user/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/bundle"

	"github.com/JailtonJunior94/devkit-go/pkg/httpserver"
)

func RegisterAuthModule(ioc *bundle.Container) []httpserver.Route {
	userRepository := repositories.NewUserRepository(ioc.DB, ioc.Telemetry)
	authUseCase := usecase.NewTokenUseCase(ioc.Config, ioc.Telemetry, ioc.Hash, ioc.Jwt, userRepository)
	authHandler := http.NewAuthHandler(ioc.Telemetry, authUseCase)

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
	userRepository := repositories.NewUserRepository(ioc.DB, ioc.Telemetry)
	createUserUseCase := usecase.NewCreateUserUseCase(ioc.Telemetry, ioc.Hash, userRepository)
	userHandler := http.NewUserHandler(ioc.Telemetry, createUserUseCase)

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
