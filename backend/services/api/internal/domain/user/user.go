package user

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidEmail     = errors.New("invalid email")
	ErrEmptyDisplayName = errors.New("display name must not be empty")
)

type User struct {
	ID          uuid.UUID
	Email       string
	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func New(email, displayName string) (*User, error) {
	email = NormalizeEmail(email)
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, ErrInvalidEmail
	}

	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return nil, ErrEmptyDisplayName
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate user id: %w", err)
	}

	now := time.Now().UTC()

	return &User{
		ID:          id,
		Email:       email,
		DisplayName: displayName,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
