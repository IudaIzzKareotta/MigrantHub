package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/server/config"
)

type Server struct {
	httpServer *http.Server
}

func New(log *slog.Logger, cfg config.HTTPConfig) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"/health",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})

	return &Server{
		httpServer: &http.Server{
			Addr:              cfg.Addr(),
			Handler:           RequestLogger(log)(mux),
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			ReadTimeout:       cfg.ReadTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
