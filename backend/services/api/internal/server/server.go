package server

import (
	"context"
	"log/slog"
	"net/http"
)

type Server struct {
	httpServer *http.Server
}

func New(log *slog.Logger) *Server {
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
			Addr:    ":8080",
			Handler: RequestLogger(log)(mux),
		},
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
