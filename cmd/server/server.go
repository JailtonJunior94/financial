package server

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jailtonjunior94/financial/pkg/bundle"
)

type ApiServe struct {
}

func NewApiServe() *ApiServe {
	return &ApiServe{}
}

func (s *ApiServe) ApiServer() {
	container := bundle.NewContainer()

	router := chi.NewRouter()
	router.Use(middleware.Heartbeat("/health"))
	router.Get("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})

	server := http.Server{
		ReadTimeout:       time.Duration(10) * time.Second,
		ReadHeaderTimeout: time.Duration(10) * time.Second,
		Handler:           router,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", container.Config.HttpServerPort))
	if err != nil {
		panic(err)
	}
	server.Serve(listener)
}
