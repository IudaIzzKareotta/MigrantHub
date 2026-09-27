package security

import (
	"errors"
	"testing"
)

func TestArgon2idHasherRoundTrip(t *testing.T) {
	h := NewArgon2idHasher()

	encoded, err := h.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	ok, err := h.Verify(encoded, "correct horse battery staple")
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if !ok {
		t.Error("Verify() = false for the correct password, want true")
	}
}

func TestArgon2idHasherWrongPassword(t *testing.T) {
	h := NewArgon2idHasher()

	encoded, err := h.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	ok, err := h.Verify(encoded, "wrong password")
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if ok {
		t.Error("Verify() = true for the wrong password, want false")
	}
}

func TestArgon2idHasherDistinctSalts(t *testing.T) {
	h := NewArgon2idHasher()

	a, err := h.Hash("same password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	b, err := h.Hash("same password")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if a == b {
		t.Error("Hash() produced identical output for two calls, want distinct salts")
	}
}

func TestArgon2idHasherInvalidHash(t *testing.T) {
	h := NewArgon2idHasher()

	_, err := h.Verify("not-a-valid-hash", "whatever")
	if !errors.Is(err, ErrInvalidHash) {
		t.Fatalf("Verify() error = %v, want %v", err, ErrInvalidHash)
	}
}
