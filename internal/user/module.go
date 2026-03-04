package user

import (
	"database/sql"

	"github.com/JailtonJunior94/devkit-go/pkg/encrypt"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"
	"github.com/jailtonjunior94/financial/configs"
	"github.com/jailtonjunior94/financial/internal/user/application/usecase"
	"github.com/jailtonjunior94/financial/internal/user/infrastructure/http"
	"github.com/jailtonjunior94/financial/internal/user/infrastructure/repositories"
	"github.com/jailtonjunior94/financial/pkg/api/httperrors"
	"github.com/jailtonjunior94/financial/pkg/api/middlewares"
	"github.com/jailtonjunior94/financial/pkg/auth"
	"github.com/jailtonjunior94/financial/pkg/observability/metrics"
)

type UserModule struct {
	UserRouter *http.UserRouter
}

func NewUserModule(db *sql.DB, cfg *configs.Config, o11y observability.Observability, jwtAdapter auth.JwtAdapter) UserModule {
	hash := encrypt.NewHashAdapter()
	errorHandler := httperrors.NewErrorHandler(o11y)

	financialMetrics := metrics.NewFinancialMetrics(o11y)
	userRepository := repositories.NewUserRepository(db, o11y, financialMetrics)

	authUseCase := usecase.NewTokenUseCase(cfg, o11y, hash, jwtAdapter, userRepository)
	createUserUseCase := usecase.NewCreateUserUseCase(o11y, hash, userRepository)
	getUserUseCase := usecase.NewGetUserUseCase(o11y, financialMetrics, userRepository)
	listUsersUseCase := usecase.NewListUsersUseCase(o11y, financialMetrics, userRepository)
	updateUserUseCase := usecase.NewUpdateUserUseCase(o11y, financialMetrics, hash, userRepository)
	deleteUserUseCase := usecase.NewDeleteUserUseCase(o11y, financialMetrics, userRepository)

	authMiddleware := middlewares.NewAuthorization(jwtAdapter, o11y, errorHandler)
	ownershipMiddleware := middlewares.NewResourceOwnership(o11y, errorHandler)

	authHandler := http.NewAuthHandler(o11y, errorHandler, authUseCase)
	userHandler := http.NewUserHandler(o11y, financialMetrics, errorHandler, createUserUseCase, getUserUseCase, listUsersUseCase, updateUserUseCase, deleteUserUseCase)
	userRouter := http.NewUserRouter(authHandler, userHandler, authMiddleware, ownershipMiddleware)

	return UserModule{UserRouter: userRouter}
}
