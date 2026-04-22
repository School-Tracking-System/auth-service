package dtos

import (
	"github.com/fercho/school-tracking/services/auth/internal/core/domain"
)

// RegisterRequest represents the input payload for user registration.
type RegisterRequest struct {
	Email     string      `json:"email" validate:"required,email"`
	Password  string      `json:"password" validate:"required,min=6"`
	FirstName string      `json:"first_name" validate:"required"`
	LastName  string      `json:"last_name" validate:"required"`
	Phone     *string     `json:"phone,omitempty"`
	Role      domain.Role `json:"role" validate:"required,oneof=admin driver guardian school_staff"`
}

// LoginRequest represents the input payload for user authentication.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshRequest represents the input payload for token refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TokenResponse represents the JWT token pair returned after login or refresh.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse represents the complete payload returned after a successful login.
type AuthResponse struct {
	User   UserResponse  `json:"user"`
	Tokens TokenResponse `json:"tokens"`
}

// UserResponse represents the public user data returned in API responses.
type UserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Phone     *string `json:"phone,omitempty"`
	Role      string  `json:"role"`
	IsActive  bool    `json:"is_active"`
}
