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

type RefreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{pool: pool}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, rt *domainauth.RefreshToken) error {
	const query = `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	db := dbtxFromContext(ctx, r.pool)

	if _, err := db.Exec(ctx, query, rt.ID, rt.UserID, rt.TokenHash, rt.ExpiresAt, rt.RevokedAt, rt.CreatedAt); err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*domainauth.RefreshToken, error) {
	const query = `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	db := dbtxFromContext(ctx, r.pool)

	var rt domainauth.RefreshToken
	err := db.QueryRow(ctx, query, tokenHash).Scan(
		&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.RevokedAt, &rt.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, appauth.ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("select refresh token: %w", err)
	}

	return &rt, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE id = $1 AND revoked_at IS NULL
	`

	db := dbtxFromContext(ctx, r.pool)

	if _, err := db.Exec(ctx, query, id); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE user_id = $1 AND revoked_at IS NULL
	`

	db := dbtxFromContext(ctx, r.pool)

	if _, err := db.Exec(ctx, query, userID); err != nil {
		return fmt.Errorf("revoke all refresh tokens for user: %w", err)
	}

	return nil
}
