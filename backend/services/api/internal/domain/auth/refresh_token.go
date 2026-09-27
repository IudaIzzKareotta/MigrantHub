package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrEmptyTokenHash = errors.New("token hash must not be empty")

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func NewRefreshToken(userID uuid.UUID, tokenHash string, ttl time.Duration) (*RefreshToken, error) {
	if tokenHash == "" {
		return nil, ErrEmptyTokenHash
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token id: %w", err)
	}

	now := time.Now().UTC()

	return &RefreshToken{
		ID:        id,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
	}, nil
}

func (r *RefreshToken) IsExpired(now time.Time) bool {
	return now.After(r.ExpiresAt)
}

func (r *RefreshToken) IsRevoked() bool {
	return r.RevokedAt != nil
}
