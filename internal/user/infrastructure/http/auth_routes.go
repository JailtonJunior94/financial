package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type (
	AuthRoutes func(authRoute *authRoute)
	authRoute  struct {
		TokenHandler func(w http.ResponseWriter, r *http.Request)
	}
)

func NewAuthRoute(router *chi.Mux, authRoutes ...AuthRoutes) *authRoute {
	route := &authRoute{}
	for _, auth := range authRoutes {
		auth(route)
	}
	route.Register(router)
	return route
}

func (u *authRoute) Register(router *chi.Mux) {
	router.Route("/api/v1/token", func(r chi.Router) {
		r.Post("/", u.TokenHandler)
	})
}

func WithTokenHandler(handler func(w http.ResponseWriter, r *http.Request)) AuthRoutes {
	return func(userRouter *authRoute) {
		userRouter.TokenHandler = handler
	}
}
