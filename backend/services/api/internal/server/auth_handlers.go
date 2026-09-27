package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/application/auth"
	appuser "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/application/user"
	domainauth "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/domain/auth"
	domainuser "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/domain/user"
)

type AuthService interface {
	Register(ctx context.Context, email, displayName, password string) (*domainuser.User, error)
	Login(ctx context.Context, email, password string) (accessToken string, accessExpiresAt time.Time, refreshToken string, err error)
	Refresh(ctx context.Context, rawToken string) (accessToken string, accessExpiresAt time.Time, newRefreshToken string, err error)
	Logout(ctx context.Context, rawToken string) error
}

type authHandler struct {
	log         *slog.Logger
	service     AuthService
	userService UserService
}

type registerRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type tokenResponse struct {
	AccessToken  string    `json:"access_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	RefreshToken string    `json:"refresh_token"`
}

func (h *authHandler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "request body must be valid JSON")
		return
	}

	u, err := h.service.Register(r.Context(), req.Email, req.DisplayName, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrInvalidEmail):
			writeError(w, http.StatusBadRequest, "invalid_email", "email is not valid")
		case errors.Is(err, domainuser.ErrEmptyDisplayName):
			writeError(w, http.StatusBadRequest, "invalid_display_name", "display name must not be empty")
		case errors.Is(err, domainauth.ErrWeakPassword):
			writeError(w, http.StatusBadRequest, "weak_password", "password must be between 8 and 72 characters")
		case errors.Is(err, appuser.ErrEmailTaken):
			writeError(w, http.StatusConflict, "email_taken", "email is already registered")
		default:
			h.log.Error("register failed", slog.Any("error", err))
			writeError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		}
		return
	}

	writeJSON(w, http.StatusCreated, newUserResponse(u))
}

func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "request body must be valid JSON")
		return
	}

	access, expiresAt, refresh, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		h.log.Error("login failed", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{AccessToken: access, ExpiresAt: expiresAt, RefreshToken: refresh})
}

func (h *authHandler) refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "request body must be valid JSON")
		return
	}

	access, expiresAt, newRefresh, err := h.service.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidRefreshToken) {
			writeError(w, http.StatusUnauthorized, "invalid_refresh_token", "refresh token is invalid or expired")
			return
		}
		h.log.Error("refresh failed", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{AccessToken: access, ExpiresAt: expiresAt, RefreshToken: newRefresh})
}

func (h *authHandler) logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "request body must be valid JSON")
		return
	}

	if err := h.service.Logout(r.Context(), req.RefreshToken); err != nil {
		h.log.Error("logout failed", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *authHandler) me(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing authentication")
		return
	}

	u, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, appuser.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		h.log.Error("get current user failed", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	writeJSON(w, http.StatusOK, newUserResponse(u))
}
