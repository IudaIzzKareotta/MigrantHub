package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	appauth "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/application/auth"
	domainauth "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/domain/auth"
)

type CredentialsRepository struct {
	pool *pgxpool.Pool
}

func NewCredentialsRepository(pool *pgxpool.Pool) *CredentialsRepository {
	return &CredentialsRepository{pool: pool}
}

func (r *CredentialsRepository) Create(ctx context.Context, c *domainauth.Credentials) error {
	const query = `
		INSERT INTO credentials (user_id, password_hash, updated_at)
		VALUES ($1, $2, $3)
	`

	db := dbtxFromContext(ctx, r.pool)

	if _, err := db.Exec(ctx, query, c.UserID, c.PasswordHash, c.UpdatedAt); err != nil {
		return fmt.Errorf("insert credentials: %w", err)
	}

	return nil
}

func (r *CredentialsRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domainauth.Credentials, error) {
	const query = `
		SELECT user_id, password_hash, updated_at
		FROM credentials
		WHERE user_id = $1
	`

	db := dbtxFromContext(ctx, r.pool)

	var c domainauth.Credentials
	err := db.QueryRow(ctx, query, userID).Scan(&c.UserID, &c.PasswordHash, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appauth.ErrCredentialsNotFound
		}
		return nil, fmt.Errorf("select credentials: %w", err)
	}

	return &c, nil
}
