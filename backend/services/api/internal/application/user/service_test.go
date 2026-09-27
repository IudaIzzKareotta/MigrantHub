package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	domainuser "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/domain/user"
)

type fakeRepository struct {
	createErr error
	byID      map[uuid.UUID]*domainuser.User
	created   *domainuser.User
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{byID: make(map[uuid.UUID]*domainuser.User)}
}

func (f *fakeRepository) Create(ctx context.Context, u *domainuser.User) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = u
	f.byID[u.ID] = u
	return nil
}

func (f *fakeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error) {
	u, ok := f.byID[id]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

func (f *fakeRepository) GetByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	for _, u := range f.byID {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, ErrNotFound
}

func TestServiceRegister(t *testing.T) {
	repo := newFakeRepository()
	svc := NewService(repo)

	u, err := svc.Register(context.Background(), "ann@example.com", "Ann")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if repo.created != u {
		t.Error("Register() did not persist the created user via the repository")
	}
}

func TestServiceRegisterInvalidEmail(t *testing.T) {
	svc := NewService(newFakeRepository())

	_, err := svc.Register(context.Background(), "not-an-email", "Ann")
	if !errors.Is(err, domainuser.ErrInvalidEmail) {
		t.Fatalf("Register() error = %v, want %v", err, domainuser.ErrInvalidEmail)
	}
}

func TestServiceRegisterEmailTaken(t *testing.T) {
	repo := newFakeRepository()
	repo.createErr = ErrEmailTaken
	svc := NewService(repo)

	_, err := svc.Register(context.Background(), "ann@example.com", "Ann")
	if !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("Register() error = %v, want %v", err, ErrEmailTaken)
	}
}

func TestServiceGetByIDNotFound(t *testing.T) {
	svc := NewService(newFakeRepository())

	_, err := svc.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByID() error = %v, want %v", err, ErrNotFound)
	}
}
