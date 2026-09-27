package auth

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		plain   string
		wantErr error
	}{
		{name: "minimum length", plain: "12345678", wantErr: nil},
		{name: "too short", plain: "1234567", wantErr: ErrWeakPassword},
		{name: "maximum length", plain: stringOfLength(72), wantErr: nil},
		{name: "too long", plain: stringOfLength(73), wantErr: ErrWeakPassword},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.plain)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidatePassword() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func stringOfLength(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

func TestNewCredentials(t *testing.T) {
	userID := uuid.New()

	c, err := NewCredentials(userID, "encoded-hash")
	if err != nil {
		t.Fatalf("NewCredentials() error = %v", err)
	}
	if c.UserID != userID {
		t.Errorf("UserID = %v, want %v", c.UserID, userID)
	}
	if c.PasswordHash != "encoded-hash" {
		t.Errorf("PasswordHash = %q, want %q", c.PasswordHash, "encoded-hash")
	}
	if c.UpdatedAt.IsZero() {
		t.Error("NewCredentials() left UpdatedAt zero-valued")
	}
}

func TestNewCredentialsEmptyHash(t *testing.T) {
	_, err := NewCredentials(uuid.New(), "")
	if !errors.Is(err, ErrEmptyPasswordHash) {
		t.Fatalf("NewCredentials() error = %v, want %v", err, ErrEmptyPasswordHash)
	}
}
