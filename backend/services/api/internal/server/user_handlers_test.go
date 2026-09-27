package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	appuser "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/application/user"
	domainuser "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/domain/user"
)

type fakeUserService struct {
	registerFn func(ctx context.Context, email, displayName string) (*domainuser.User, error)
	getByIDFn  func(ctx context.Context, id uuid.UUID) (*domainuser.User, error)
}

func (f *fakeUserService) Register(ctx context.Context, email, displayName string) (*domainuser.User, error) {
	return f.registerFn(ctx, email, displayName)
}

func (f *fakeUserService) GetByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error) {
	return f.getByIDFn(ctx, id)
}

func newTestRouter(svc UserService) http.Handler {
	h := &userHandler{log: slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil)), service: svc}

	r := chi.NewRouter()
	r.Route("/users", func(r chi.Router) {
		r.Post("/", h.create)
		r.Get("/{id}", h.getByID)
	})
	return r
}

func TestUserHandlerCreate(t *testing.T) {
	fixedUser, _ := domainuser.New("ann@example.com", "Ann")

	tests := []struct {
		name       string
		body       string
		registerFn func(ctx context.Context, email, displayName string) (*domainuser.User, error)
		wantStatus int
		wantCode   string
	}{
		{
			name: "success",
			body: `{"email":"ann@example.com","display_name":"Ann"}`,
			registerFn: func(ctx context.Context, email, displayName string) (*domainuser.User, error) {
				return fixedUser, nil
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid json body",
			body:       `not-json`,
			registerFn: nil,
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_body",
		},
		{
			name: "invalid email",
			body: `{"email":"nope","display_name":"Ann"}`,
			registerFn: func(ctx context.Context, email, displayName string) (*domainuser.User, error) {
				return nil, domainuser.ErrInvalidEmail
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_email",
		},
		{
			name: "email taken",
			body: `{"email":"ann@example.com","display_name":"Ann"}`,
			registerFn: func(ctx context.Context, email, displayName string) (*domainuser.User, error) {
				return nil, appuser.ErrEmailTaken
			},
			wantStatus: http.StatusConflict,
			wantCode:   "email_taken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newTestRouter(&fakeUserService{registerFn: tt.registerFn})

			req := httptest.NewRequest(http.MethodPost, "/users/", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantCode != "" {
				var got errorResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("decode error response: %v", err)
				}
				if got.Error.Code != tt.wantCode {
					t.Errorf("error code = %q, want %q", got.Error.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestUserHandlerGetByID(t *testing.T) {
	fixedUser, _ := domainuser.New("ann@example.com", "Ann")

	tests := []struct {
		name       string
		path       string
		getByIDFn  func(ctx context.Context, id uuid.UUID) (*domainuser.User, error)
		wantStatus int
		wantCode   string
	}{
		{
			name: "success",
			path: "/users/" + fixedUser.ID.String(),
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*domainuser.User, error) {
				return fixedUser, nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid uuid",
			path:       "/users/not-a-uuid",
			getByIDFn:  nil,
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_id",
		},
		{
			name: "not found",
			path: "/users/" + uuid.New().String(),
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*domainuser.User, error) {
				return nil, appuser.ErrNotFound
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "not_found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newTestRouter(&fakeUserService{getByIDFn: tt.getByIDFn})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}

			if tt.wantCode != "" {
				var got errorResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("decode error response: %v", err)
				}
				if got.Error.Code != tt.wantCode {
					t.Errorf("error code = %q, want %q", got.Error.Code, tt.wantCode)
				}
			}
		})
	}
}
