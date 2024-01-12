package web

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	http *http.Server
}

func NewServer(port string, handler http.Handler) *Server {
	return &Server{
		http: &http.Server{
			Addr:           ":" + port,
			Handler:        handler,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20, // 1mb
		},
	}
}

func (s *Server) Start() error {
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
