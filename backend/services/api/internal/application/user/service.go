package user

import (
	"context"
	"errors"

	"github.com/google/uuid"

	domainuser "github.com/IudaIzzKareotta/MigrantHub/backend/services/api/internal/domain/user"
)

var (
	ErrNotFound   = errors.New("user not found")
	ErrEmailTaken = errors.New("email already registered")
)

type Repository interface {
	Create(ctx context.Context, u *domainuser.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, email, displayName string) (*domainuser.User, error) {
	u, err := domainuser.New(email, displayName)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error) {
	return s.repo.GetByID(ctx, id)
}
