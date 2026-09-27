package auth

import (
	"errors"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 72
)

var (
	ErrWeakPassword      = errors.New("password must be between 8 and 72 characters")
	ErrEmptyPasswordHash = errors.New("password hash must not be empty")
)

func ValidatePassword(plain string) error {
	n := utf8.RuneCountInString(plain)
	if n < minPasswordLength || n > maxPasswordLength {
		return ErrWeakPassword
	}
	return nil
}

type Credentials struct {
	UserID       uuid.UUID
	PasswordHash string
	UpdatedAt    time.Time
}

func NewCredentials(userID uuid.UUID, passwordHash string) (*Credentials, error) {
	if passwordHash == "" {
		return nil, ErrEmptyPasswordHash
	}

	return &Credentials{
		UserID:       userID,
		PasswordHash: passwordHash,
		UpdatedAt:    time.Now().UTC(),
	}, nil
}
