package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/server/config"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type Server struct {
	httpServer *http.Server
}

func New(log *slog.Logger, cfg config.HTTPConfig, db Pinger, userService UserService, authService AuthService, tokenVerifier TokenVerifier) *Server {
	r := chi.NewRouter()
	r.Use(RequestLogger(log))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	users := &userHandler{log: log, service: userService}
	r.Route("/users", func(r chi.Router) {
		r.Post("/", users.create)
		r.Get("/{id}", users.getByID)
	})

	authH := &authHandler{log: log, service: authService, userService: userService}
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authH.register)
		r.Post("/login", authH.login)
		r.Post("/refresh", authH.refresh)
		r.Post("/logout", authH.logout)

		r.Group(func(r chi.Router) {
			r.Use(RequireAuth(tokenVerifier))
			r.Get("/me", authH.me)
		})
	})

	return &Server{
		httpServer: &http.Server{
			Addr:              cfg.Addr(),
			Handler:           r,
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
