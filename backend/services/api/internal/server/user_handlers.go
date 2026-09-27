package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	appuser "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/application/user"
	domainuser "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/domain/user"
)

type UserService interface {
	Register(ctx context.Context, email, displayName string) (*domainuser.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error)
}

type userHandler struct {
	log     *slog.Logger
	service UserService
}

type createUserRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

type userResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func newUserResponse(u *domainuser.User) userResponse {
	return userResponse{
		ID:          u.ID.String(),
		Email:       u.Email,
		DisplayName: u.DisplayName,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

func (h *userHandler) create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "request body must be valid JSON")
		return
	}

	u, err := h.service.Register(r.Context(), req.Email, req.DisplayName)
	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrInvalidEmail):
			writeError(w, http.StatusBadRequest, "invalid_email", "email is not valid")
		case errors.Is(err, domainuser.ErrEmptyDisplayName):
			writeError(w, http.StatusBadRequest, "invalid_display_name", "display name must not be empty")
		case errors.Is(err, appuser.ErrEmailTaken):
			writeError(w, http.StatusConflict, "email_taken", "email is already registered")
		default:
			h.log.Error("register user failed", slog.Any("error", err))
			writeError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		}
		return
	}

	writeJSON(w, http.StatusCreated, newUserResponse(u))
}

func (h *userHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "id must be a valid UUID")
		return
	}

	u, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, appuser.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "user not found")
			return
		}
		h.log.Error("get user failed", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "something went wrong")
		return
	}

	writeJSON(w, http.StatusOK, newUserResponse(u))
}
