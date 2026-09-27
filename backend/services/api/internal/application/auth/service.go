package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	domainauth "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/domain/auth"
	domainuser "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/domain/user"

	userapp "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/application/user"
)

var (
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrCredentialsNotFound  = errors.New("credentials not found")
	ErrInvalidRefreshToken  = errors.New("invalid refresh token")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
)

type CredentialsRepository interface {
	Create(ctx context.Context, c *domainauth.Credentials) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domainauth.Credentials, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *domainauth.RefreshToken) error
	GetByHash(ctx context.Context, tokenHash string) (*domainauth.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

type PasswordHasher interface {
	Hash(plain string) (string, error)
	Verify(hash, plain string) (bool, error)
}

type AccessTokenIssuer interface {
	Issue(userID uuid.UUID) (token string, expiresAt time.Time, err error)
}

type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type UserRegistrar interface {
	Register(ctx context.Context, email, displayName string) (*domainuser.User, error)
	GetByEmail(ctx context.Context, email string) (*domainuser.User, error)
}

type Service struct {
	users         UserRegistrar
	credentials   CredentialsRepository
	refreshTokens RefreshTokenRepository
	hasher        PasswordHasher
	tokens        AccessTokenIssuer
	tx            Transactor
	refreshTTL    time.Duration
}

func NewService(
	users UserRegistrar,
	credentials CredentialsRepository,
	refreshTokens RefreshTokenRepository,
	hasher PasswordHasher,
	tokens AccessTokenIssuer,
	tx Transactor,
	refreshTTL time.Duration,
) *Service {
	return &Service{
		users:         users,
		credentials:   credentials,
		refreshTokens: refreshTokens,
		hasher:        hasher,
		tokens:        tokens,
		tx:            tx,
		refreshTTL:    refreshTTL,
	}
}

func (s *Service) Register(ctx context.Context, email, displayName, password string) (*domainuser.User, error) {
	if err := domainauth.ValidatePassword(password); err != nil {
		return nil, err
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	var u *domainuser.User
	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		u, err = s.users.Register(ctx, email, displayName)
		if err != nil {
			return err
		}

		creds, err := domainauth.NewCredentials(u.ID, hash)
		if err != nil {
			return err
		}

		return s.credentials.Create(ctx, creds)
	})
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (accessToken string, accessExpiresAt time.Time, refreshToken string, err error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, userapp.ErrNotFound) {
			return "", time.Time{}, "", ErrInvalidCredentials
		}
		return "", time.Time{}, "", fmt.Errorf("get user by email: %w", err)
	}

	creds, err := s.credentials.GetByUserID(ctx, u.ID)
	if err != nil {
		if errors.Is(err, ErrCredentialsNotFound) {
			return "", time.Time{}, "", ErrInvalidCredentials
		}
		return "", time.Time{}, "", fmt.Errorf("get credentials: %w", err)
	}

	ok, err := s.hasher.Verify(creds.PasswordHash, password)
	if err != nil {
		return "", time.Time{}, "", fmt.Errorf("verify password: %w", err)
	}
	if !ok {
		return "", time.Time{}, "", ErrInvalidCredentials
	}

	accessToken, accessExpiresAt, err = s.tokens.Issue(u.ID)
	if err != nil {
		return "", time.Time{}, "", fmt.Errorf("issue access token: %w", err)
	}

	refreshToken, err = s.issueRefreshToken(ctx, u.ID)
	if err != nil {
		return "", time.Time{}, "", err
	}

	return accessToken, accessExpiresAt, refreshToken, nil
}

func (s *Service) Refresh(ctx context.Context, rawToken string) (accessToken string, accessExpiresAt time.Time, newRefreshToken string, err error) {
	hash := hashToken(rawToken)

	rt, err := s.refreshTokens.GetByHash(ctx, hash)
	if err != nil {
		return "", time.Time{}, "", ErrInvalidRefreshToken
	}

	if rt.IsRevoked() {
		_ = s.refreshTokens.RevokeAllForUser(ctx, rt.UserID)
		return "", time.Time{}, "", ErrInvalidRefreshToken
	}

	if rt.IsExpired(time.Now().UTC()) {
		return "", time.Time{}, "", ErrInvalidRefreshToken
	}

	if err := s.refreshTokens.Revoke(ctx, rt.ID); err != nil {
		return "", time.Time{}, "", fmt.Errorf("revoke old refresh token: %w", err)
	}

	accessToken, accessExpiresAt, err = s.tokens.Issue(rt.UserID)
	if err != nil {
		return "", time.Time{}, "", fmt.Errorf("issue access token: %w", err)
	}

	newRefreshToken, err = s.issueRefreshToken(ctx, rt.UserID)
	if err != nil {
		return "", time.Time{}, "", err
	}

	return accessToken, accessExpiresAt, newRefreshToken, nil
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	rt, err := s.refreshTokens.GetByHash(ctx, hashToken(rawToken))
	if err != nil {
		return nil
	}

	return s.refreshTokens.Revoke(ctx, rt.ID)
}

func (s *Service) issueRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	raw, err := randomToken(32)
	if err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}

	rt, err := domainauth.NewRefreshToken(userID, hashToken(raw), s.refreshTTL)
	if err != nil {
		return "", err
	}

	if err := s.refreshTokens.Create(ctx, rt); err != nil {
		return "", fmt.Errorf("store refresh token: %w", err)
	}

	return raw, nil
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
