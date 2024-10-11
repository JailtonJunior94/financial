package http

import (
	"github.com/JailtonJunior94/devkit-go/pkg/httpserver"
)

type authRoutes struct {
	routes []httpserver.Route
}

func NewAuthRoutes() *authRoutes {
	return &authRoutes{}
}

func (u *authRoutes) Register(route httpserver.Route) {
	u.routes = append(u.routes, route)
}

func (u *authRoutes) Routes() []httpserver.Route {
	return u.routes
}
