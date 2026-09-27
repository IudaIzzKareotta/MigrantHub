package user

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		displayName string
		wantErr     error
	}{
		{
			name:        "valid input",
			email:       "  Ann@Example.com ",
			displayName: "  Ann  ",
			wantErr:     nil,
		},
		{
			name:        "invalid email",
			email:       "not-an-email",
			displayName: "Ann",
			wantErr:     ErrInvalidEmail,
		},
		{
			name:        "empty email",
			email:       "",
			displayName: "Ann",
			wantErr:     ErrInvalidEmail,
		},
		{
			name:        "empty display name",
			email:       "ann@example.com",
			displayName: "   ",
			wantErr:     ErrEmptyDisplayName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := New(tt.email, tt.displayName)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("New() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("New() error = %v, want nil", err)
			}
			if u.ID == uuid.Nil {
				t.Error("New() left ID zero-valued")
			}
			if u.Email != "ann@example.com" {
				t.Errorf("Email = %q, want normalized %q", u.Email, "ann@example.com")
			}
			if u.DisplayName != "Ann" {
				t.Errorf("DisplayName = %q, want trimmed %q", u.DisplayName, "Ann")
			}
			if u.CreatedAt.IsZero() || u.UpdatedAt.IsZero() {
				t.Error("New() left timestamps zero-valued")
			}
			if !u.CreatedAt.Equal(u.UpdatedAt) {
				t.Errorf("CreatedAt (%v) != UpdatedAt (%v) on creation", u.CreatedAt, u.UpdatedAt)
			}
		})
	}
}
