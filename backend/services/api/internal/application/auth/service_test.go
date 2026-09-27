package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	userapp "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/application/user"
	domainauth "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/domain/auth"
	domainuser "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/domain/user"
)

type fakeUsers struct {
	byEmail    map[string]*domainuser.User
	registerFn func(ctx context.Context, email, displayName string) (*domainuser.User, error)
}

func newFakeUsers() *fakeUsers {
	return &fakeUsers{byEmail: make(map[string]*domainuser.User)}
}

func (f *fakeUsers) Register(ctx context.Context, email, displayName string) (*domainuser.User, error) {
	if f.registerFn != nil {
		return f.registerFn(ctx, email, displayName)
	}
	u, err := domainuser.New(email, displayName)
	if err != nil {
		return nil, err
	}
	f.byEmail[u.Email] = u
	return u, nil
}

func (f *fakeUsers) GetByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	u, ok := f.byEmail[domainuser.NormalizeEmail(email)]
	if !ok {
		return nil, userapp.ErrNotFound
	}
	return u, nil
}

type fakeCredentials struct {
	byUserID map[uuid.UUID]*domainauth.Credentials
}

func newFakeCredentials() *fakeCredentials {
	return &fakeCredentials{byUserID: make(map[uuid.UUID]*domainauth.Credentials)}
}

func (f *fakeCredentials) Create(ctx context.Context, c *domainauth.Credentials) error {
	f.byUserID[c.UserID] = c
	return nil
}

func (f *fakeCredentials) GetByUserID(ctx context.Context, userID uuid.UUID) (*domainauth.Credentials, error) {
	c, ok := f.byUserID[userID]
	if !ok {
		return nil, ErrCredentialsNotFound
	}
	return c, nil
}

type fakeRefreshTokens struct {
	byHash map[string]*domainauth.RefreshToken
}

func newFakeRefreshTokens() *fakeRefreshTokens {
	return &fakeRefreshTokens{byHash: make(map[string]*domainauth.RefreshToken)}
}

func (f *fakeRefreshTokens) Create(ctx context.Context, rt *domainauth.RefreshToken) error {
	f.byHash[rt.TokenHash] = rt
	return nil
}

func (f *fakeRefreshTokens) GetByHash(ctx context.Context, tokenHash string) (*domainauth.RefreshToken, error) {
	rt, ok := f.byHash[tokenHash]
	if !ok {
		return nil, ErrRefreshTokenNotFound
	}
	return rt, nil
}

func (f *fakeRefreshTokens) Revoke(ctx context.Context, id uuid.UUID) error {
	for _, rt := range f.byHash {
		if rt.ID == id {
			now := time.Now().UTC()
			rt.RevokedAt = &now
		}
	}
	return nil
}

func (f *fakeRefreshTokens) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	for _, rt := range f.byHash {
		if rt.UserID == userID {
			now := time.Now().UTC()
			rt.RevokedAt = &now
		}
	}
	return nil
}

type plaintextHasher struct{}

func (plaintextHasher) Hash(plain string) (string, error) { return "hashed:" + plain, nil }
func (plaintextHasher) Verify(hash, plain string) (bool, error) {
	return hash == "hashed:"+plain, nil
}

type fakeTokens struct{}

func (fakeTokens) Issue(userID uuid.UUID) (string, time.Time, error) {
	return "access-for-" + userID.String(), time.Now().Add(time.Hour), nil
}

type noopTransactor struct{}

func (noopTransactor) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func newTestService() (*Service, *fakeUsers, *fakeCredentials, *fakeRefreshTokens) {
	users := newFakeUsers()
	credentials := newFakeCredentials()
	refreshTokens := newFakeRefreshTokens()

	svc := NewService(users, credentials, refreshTokens, plaintextHasher{}, fakeTokens{}, noopTransactor{}, time.Hour)

	return svc, users, credentials, refreshTokens
}

func TestServiceRegister(t *testing.T) {
	svc, _, credentials, _ := newTestService()

	u, err := svc.Register(context.Background(), "ann@example.com", "Ann", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	creds, ok := credentials.byUserID[u.ID]
	if !ok {
		t.Fatal("Register() did not create credentials for the new user")
	}
	if creds.PasswordHash != "hashed:password123" {
		t.Errorf("PasswordHash = %q, want %q", creds.PasswordHash, "hashed:password123")
	}
}

func TestServiceRegisterWeakPassword(t *testing.T) {
	svc, _, _, _ := newTestService()

	_, err := svc.Register(context.Background(), "ann@example.com", "Ann", "short")
	if !errors.Is(err, domainauth.ErrWeakPassword) {
		t.Fatalf("Register() error = %v, want %v", err, domainauth.ErrWeakPassword)
	}
}

func TestServiceRegisterRollsBackOnUserFailure(t *testing.T) {
	svc, _, credentials, _ := newTestService()
	wantErr := errors.New("boom")
	svc.users.(*fakeUsers).registerFn = func(ctx context.Context, email, displayName string) (*domainuser.User, error) {
		return nil, wantErr
	}

	_, err := svc.Register(context.Background(), "ann@example.com", "Ann", "password123")
	if !errors.Is(err, wantErr) {
		t.Fatalf("Register() error = %v, want %v", err, wantErr)
	}
	if len(credentials.byUserID) != 0 {
		t.Error("Register() left credentials behind despite user creation failing")
	}
}

func TestServiceLoginSuccess(t *testing.T) {
	svc, _, _, _ := newTestService()

	u, err := svc.Register(context.Background(), "ann@example.com", "Ann", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	access, expiresAt, refresh, err := svc.Login(context.Background(), "ann@example.com", "password123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if access != "access-for-"+u.ID.String() {
		t.Errorf("access token = %q, want it to encode user id %v", access, u.ID)
	}
	if !expiresAt.After(time.Now()) {
		t.Error("Login() returned an already-expired access token")
	}
	if refresh == "" {
		t.Error("Login() returned an empty refresh token")
	}
}

func TestServiceLoginWrongPassword(t *testing.T) {
	svc, _, _, _ := newTestService()

	if _, err := svc.Register(context.Background(), "ann@example.com", "Ann", "password123"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	_, _, _, err := svc.Login(context.Background(), "ann@example.com", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestServiceLoginUnknownEmail(t *testing.T) {
	svc, _, _, _ := newTestService()

	_, _, _, err := svc.Login(context.Background(), "nobody@example.com", "password123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestServiceRefreshRotatesToken(t *testing.T) {
	svc, _, _, refreshTokens := newTestService()

	if _, err := svc.Register(context.Background(), "ann@example.com", "Ann", "password123"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	_, _, refresh, err := svc.Login(context.Background(), "ann@example.com", "password123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	_, _, newRefresh, err := svc.Refresh(context.Background(), refresh)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if newRefresh == refresh {
		t.Error("Refresh() returned the same refresh token, want a rotated one")
	}

	oldHash := hashToken(refresh)
	oldToken, ok := refreshTokens.byHash[oldHash]
	if !ok {
		t.Fatal("old refresh token disappeared from the store")
	}
	if !oldToken.IsRevoked() {
		t.Error("Refresh() did not revoke the old refresh token")
	}
}

func TestServiceRefreshReuseOfRevokedTokenRevokesAllSessions(t *testing.T) {
	svc, _, _, refreshTokens := newTestService()

	if _, err := svc.Register(context.Background(), "ann@example.com", "Ann", "password123"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	_, _, refresh, err := svc.Login(context.Background(), "ann@example.com", "password123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	_, _, secondRefresh, err := svc.Refresh(context.Background(), refresh)
	if err != nil {
		t.Fatalf("first Refresh() error = %v", err)
	}

	_, _, _, err = svc.Refresh(context.Background(), refresh)
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("reuse Refresh() error = %v, want %v", err, ErrInvalidRefreshToken)
	}

	secondHash := hashToken(secondRefresh)
	secondToken, ok := refreshTokens.byHash[secondHash]
	if !ok {
		t.Fatal("second refresh token disappeared from the store")
	}
	if !secondToken.IsRevoked() {
		t.Error("reuse of a revoked refresh token should revoke all sessions, but the second token is still active")
	}
}

func TestServiceRefreshInvalidToken(t *testing.T) {
	svc, _, _, _ := newTestService()

	_, _, _, err := svc.Refresh(context.Background(), "does-not-exist")
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("Refresh() error = %v, want %v", err, ErrInvalidRefreshToken)
	}
}

func TestServiceLogout(t *testing.T) {
	svc, _, _, refreshTokens := newTestService()

	if _, err := svc.Register(context.Background(), "ann@example.com", "Ann", "password123"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	_, _, refresh, err := svc.Login(context.Background(), "ann@example.com", "password123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if err := svc.Logout(context.Background(), refresh); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	rt, ok := refreshTokens.byHash[hashToken(refresh)]
	if !ok {
		t.Fatal("refresh token disappeared from the store")
	}
	if !rt.IsRevoked() {
		t.Error("Logout() did not revoke the refresh token")
	}
}

func TestServiceLogoutUnknownTokenIsIdempotent(t *testing.T) {
	svc, _, _, _ := newTestService()

	if err := svc.Logout(context.Background(), "does-not-exist"); err != nil {
		t.Fatalf("Logout() error = %v, want nil for an unknown token", err)
	}
}
