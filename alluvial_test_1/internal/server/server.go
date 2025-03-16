package server

import (
	"context"
	"fmt"
	"net/http"
)

type Server struct {
	server *http.Server
}

func New(handler http.Handler, port string) *Server {
	return &Server{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: handler,
		},
	}
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
