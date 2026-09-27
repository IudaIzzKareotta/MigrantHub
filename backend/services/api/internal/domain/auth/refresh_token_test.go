package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewRefreshToken(t *testing.T) {
	userID := uuid.New()

	rt, err := NewRefreshToken(userID, "hash", time.Hour)
	if err != nil {
		t.Fatalf("NewRefreshToken() error = %v", err)
	}

	if rt.ID == uuid.Nil {
		t.Error("NewRefreshToken() left ID zero-valued")
	}
	if rt.UserID != userID {
		t.Errorf("UserID = %v, want %v", rt.UserID, userID)
	}
	if rt.TokenHash != "hash" {
		t.Errorf("TokenHash = %q, want %q", rt.TokenHash, "hash")
	}
	if rt.IsRevoked() {
		t.Error("new refresh token should not be revoked")
	}
	if !rt.ExpiresAt.After(rt.CreatedAt) {
		t.Errorf("ExpiresAt (%v) should be after CreatedAt (%v)", rt.ExpiresAt, rt.CreatedAt)
	}
}

func TestNewRefreshTokenEmptyHash(t *testing.T) {
	_, err := NewRefreshToken(uuid.New(), "", time.Hour)
	if !errors.Is(err, ErrEmptyTokenHash) {
		t.Fatalf("NewRefreshToken() error = %v, want %v", err, ErrEmptyTokenHash)
	}
}

func TestRefreshTokenIsExpired(t *testing.T) {
	rt, err := NewRefreshToken(uuid.New(), "hash", time.Hour)
	if err != nil {
		t.Fatalf("NewRefreshToken() error = %v", err)
	}

	if rt.IsExpired(rt.CreatedAt) {
		t.Error("IsExpired() = true right after creation, want false")
	}
	if !rt.IsExpired(rt.ExpiresAt.Add(time.Second)) {
		t.Error("IsExpired() = false after expiry, want true")
	}
}

func TestRefreshTokenIsRevoked(t *testing.T) {
	rt, err := NewRefreshToken(uuid.New(), "hash", time.Hour)
	if err != nil {
		t.Fatalf("NewRefreshToken() error = %v", err)
	}

	if rt.IsRevoked() {
		t.Error("IsRevoked() = true before revocation, want false")
	}

	now := time.Now().UTC()
	rt.RevokedAt = &now

	if !rt.IsRevoked() {
		t.Error("IsRevoked() = false after revocation, want true")
	}
}
