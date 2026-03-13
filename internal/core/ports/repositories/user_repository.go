package repositories

import (
	"context"

	"github.com/fercho/school-tracking/services/auth/internal/core/domain"
	"github.com/google/uuid"
)

// UserRepository defines the persistence contract for user entities.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}
