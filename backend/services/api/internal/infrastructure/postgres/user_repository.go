package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	appuser "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/application/user"
	domainuser "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/domain/user"
)

const uniqueViolationCode = "23505"

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *domainuser.User) error {
	const query = `
		INSERT INTO users (id, email, display_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	db := dbtxFromContext(ctx, r.pool)

	if _, err := db.Exec(ctx, query, u.ID, u.Email, u.DisplayName, u.CreatedAt, u.UpdatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return appuser.ErrEmailTaken
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error) {
	const query = `
		SELECT id, email, display_name, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	db := dbtxFromContext(ctx, r.pool)

	var u domainuser.User
	err := db.QueryRow(ctx, query, id).Scan(&u.ID, &u.Email, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appuser.ErrNotFound
		}
		return nil, fmt.Errorf("select user: %w", err)
	}

	return &u, nil
}
