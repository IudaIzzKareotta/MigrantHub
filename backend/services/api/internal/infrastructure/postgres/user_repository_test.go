package postgres

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	appuser "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/application/user"
	domainuser "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/domain/user"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping postgres integration test")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool
}

func txContext(t *testing.T, pool *pgxpool.Pool) context.Context {
	t.Helper()

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback(ctx)
	})

	return context.WithValue(ctx, txKey{}, tx)
}

func uniqueTestEmail(t *testing.T) string {
	t.Helper()
	return uuid.NewString() + "@example.com"
}

func TestUserRepositoryCreateAndGetByID(t *testing.T) {
	pool := testPool(t)
	ctx := txContext(t, pool)
	repo := NewUserRepository(pool)

	u, err := domainuser.New(uniqueTestEmail(t), "Ann")
	if err != nil {
		t.Fatalf("domainuser.New() error = %v", err)
	}

	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if got.ID != u.ID {
		t.Errorf("ID = %v, want %v", got.ID, u.ID)
	}
	if got.Email != u.Email {
		t.Errorf("Email = %q, want %q", got.Email, u.Email)
	}
	if got.DisplayName != u.DisplayName {
		t.Errorf("DisplayName = %q, want %q", got.DisplayName, u.DisplayName)
	}
	if !got.CreatedAt.Equal(u.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, u.CreatedAt)
	}
}

func TestUserRepositoryCreateDuplicateEmail(t *testing.T) {
	pool := testPool(t)
	ctx := txContext(t, pool)
	repo := NewUserRepository(pool)

	email := uniqueTestEmail(t)

	u1, _ := domainuser.New(email, "First")
	if err := repo.Create(ctx, u1); err != nil {
		t.Fatalf("Create() first user error = %v", err)
	}

	u2, _ := domainuser.New(email, "Second")
	err := repo.Create(ctx, u2)
	if !errors.Is(err, appuser.ErrEmailTaken) {
		t.Fatalf("Create() error = %v, want %v", err, appuser.ErrEmailTaken)
	}
}

func TestUserRepositoryGetByIDNotFound(t *testing.T) {
	pool := testPool(t)
	ctx := txContext(t, pool)
	repo := NewUserRepository(pool)

	_, err := repo.GetByID(ctx, uuid.New())
	if !errors.Is(err, appuser.ErrNotFound) {
		t.Fatalf("GetByID() error = %v, want %v", err, appuser.ErrNotFound)
	}
}
