package httpserver

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	http *http.Server
}

type Config struct {
	// Port http server port
	Port int

	// ReadTimeout read timeout
	ReadTimeout time.Duration

	// WriteTimeout write timeout
	WriteTimeout time.Duration

	// MaxHeaderSizeMb max size of headers in megabytes
	MaxHeaderSizeMb int
}

func NewServer(handler http.Handler, cfg Config) *Server {
	p := strconv.Itoa(cfg.Port)
	return &Server{
		http: &http.Server{
			Addr:           ":" + p,
			Handler:        handler,
			ReadTimeout:    cfg.ReadTimeout,
			WriteTimeout:   cfg.WriteTimeout,
			MaxHeaderBytes: cfg.MaxHeaderSizeMb << 20,
		},
	}
}

// NewDefaultServer returns http server with default config
func NewDefaultServer(handler http.Handler, port int) *Server {
	p := strconv.Itoa(port)
	return &Server{
		http: &http.Server{
			Addr:           ":" + p,
			Handler:        handler,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}
}

func (s *Server) Start() error {
	return s.http.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
