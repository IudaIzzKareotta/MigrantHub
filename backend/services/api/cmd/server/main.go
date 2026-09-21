package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IudaIzzKareotta/MigrantHub/backend/pkg/logger"
	"github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/server"
)

func main() {
	log := logger.New(logger.Config{
		// Environment: os.Getenv("APP_ENV"),
		Environment: "dev",
	})

	srv := server.New(log)

	srvErr := make(chan error, 1)

	go func() {
		log.Info("starting HTTP server",
			slog.String("service", "api"),
			slog.String("addr", ":8080"),
		)

		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErr <- err
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-srvErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed",
				slog.Any("error", err),
			)

			os.Exit(1)
		}

	case sig := <-sig:
		log.Info("shutdown signal received",
			slog.String("signal", sig.String()),
		)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server shutdown failed",
			slog.Any("error", err),
		)

		os.Exit(1)
	}

	log.Info("server stopped")
}
