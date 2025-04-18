package httpserver

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	http *http.Server
}

type Config struct {
	// Port http server port
	Port string

	// ReadTimeout read timeout
	ReadTimeout time.Duration

	// WriteTimeout write timeout
	WriteTimeout time.Duration

	// MaxHeaderSizeMb max size of headers in megabytes
	MaxHeaderSizeMb int
}

func NewServer(handler http.Handler, cfg Config) *Server {
	return &Server{
		http: &http.Server{
			Addr:           ":" + cfg.Port,
			Handler:        handler,
			ReadTimeout:    cfg.ReadTimeout,
			WriteTimeout:   cfg.WriteTimeout,
			MaxHeaderBytes: cfg.MaxHeaderSizeMb << 20,
		},
	}
}

func (s *Server) Start() error {
	return s.http.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
