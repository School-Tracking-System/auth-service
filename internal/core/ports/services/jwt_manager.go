package services

import (
	"github.com/fercho/school-tracking/services/auth/internal/core/domain"
)

// JWTManager defines the contract for generating and validating JWT tokens.
type JWTManager interface {
	Generate(user *domain.User) (*domain.TokenPair, error)
	ValidateAccessToken(token string) (*domain.Claims, error)
	ValidateRefreshToken(token string) (*domain.Claims, error)
}
