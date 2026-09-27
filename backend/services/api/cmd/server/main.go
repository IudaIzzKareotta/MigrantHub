package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IudaIzzKareotta/MigrantHub/backend/pkg/logger"
	authapp "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/application/auth"
	userapp "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/application/user"
	"github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/infrastructure/jwtauth"
	"github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/infrastructure/postgres"
	"github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/infrastructure/security"
	"github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/server"
	"github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/server/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(logger.Config{
		Environment: cfg.App.Environment,
	})

	pool, err := postgres.NewPool(context.Background(), cfg.DB)
	if err != nil {
		log.Error("connect to database failed",
			slog.Any("error", err),
		)

		os.Exit(1)
	}
	defer pool.Close()

	userRepo := postgres.NewUserRepository(pool)
	userService := userapp.NewService(userRepo)

	transactor := postgres.NewTransactor(pool)
	credentialsRepo := postgres.NewCredentialsRepository(pool)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(pool)
	hasher := security.NewArgon2idHasher()
	tokenIssuer := jwtauth.NewIssuer(cfg.Auth.JWTSecret, cfg.Auth.AccessTokenTTL)

	authService := authapp.NewService(
		userService,
		credentialsRepo,
		refreshTokenRepo,
		hasher,
		tokenIssuer,
		transactor,
		cfg.Auth.RefreshTokenTTL,
	)

	srv := server.New(log, cfg.HTTP, pool, userService, authService, tokenIssuer)

	srvErr := make(chan error, 1)

	go func() {
		log.Info("starting HTTP server",
			slog.String("service", "api"),
			slog.String("addr", cfg.HTTP.Addr()),
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
