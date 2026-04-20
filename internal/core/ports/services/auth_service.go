package services

import (
	"context"

	"github.com/fercho/school-tracking/services/auth/internal/core/domain"
)

// AuthService defines the contract for authentication and user registration operations.
type AuthService interface {
	Register(ctx context.Context, params domain.RegisterParams) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.User, *domain.TokenPair, error)	ValidateToken(ctx context.Context, token string) (*domain.Claims, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
}
