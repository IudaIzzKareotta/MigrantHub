package jwtauth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestIssuerRoundTrip(t *testing.T) {
	issuer := NewIssuer("a-32-plus-character-test-secret!", time.Hour)
	userID := uuid.New()

	token, expiresAt, err := issuer.Issue(userID)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if !expiresAt.After(time.Now().UTC()) {
		t.Errorf("expiresAt = %v, want it in the future", expiresAt)
	}

	got, err := issuer.Verify(token)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if got != userID {
		t.Errorf("Verify() = %v, want %v", got, userID)
	}
}

func TestIssuerVerifyExpired(t *testing.T) {
	issuer := NewIssuer("a-32-plus-character-test-secret!", -time.Minute)

	token, _, err := issuer.Issue(uuid.New())
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	_, err = issuer.Verify(token)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Verify() error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestIssuerVerifyWrongSecret(t *testing.T) {
	issuer := NewIssuer("a-32-plus-character-test-secret!", time.Hour)
	other := NewIssuer("a-different-32-plus-char-secret", time.Hour)

	token, _, err := issuer.Issue(uuid.New())
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	_, err = other.Verify(token)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Verify() error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestIssuerVerifyMalformedToken(t *testing.T) {
	issuer := NewIssuer("a-32-plus-character-test-secret!", time.Hour)

	_, err := issuer.Verify("not-a-jwt")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Verify() error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestIssuerVerifyRejectsAlgNone(t *testing.T) {
	issuer := NewIssuer("a-32-plus-character-test-secret!", time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{
		Subject:   uuid.New().String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})

	signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	_, err = issuer.Verify(signed)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Verify() error = %v, want %v", err, ErrInvalidToken)
	}
}
