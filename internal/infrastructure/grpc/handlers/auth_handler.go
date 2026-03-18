package handlers

import (
	"context"

	authv1 "github.com/fercho/school-tracking/proto/gen/auth/v1"
	"github.com/fercho/school-tracking/services/auth/internal/core/ports/services"
)

// AuthHandler handles gRPC requests for authentication and token validation.
type AuthHandler struct {
	authv1.UnimplementedAuthServiceServer
	authService services.AuthService
}

// NewAuthHandler creates a new AuthHandler with the given auth service.
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// ValidateToken verifies a JWT token and returns its claims.
func (h *AuthHandler) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	if req.AccessToken == "" {
		return &authv1.ValidateTokenResponse{IsValid: false}, nil
	}

	claims, err := h.authService.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		return &authv1.ValidateTokenResponse{IsValid: false}, nil
	}

	return &authv1.ValidateTokenResponse{
		IsValid: true,
		UserId:  claims.UserID,
		Role:    string(claims.Role),
		Email:   claims.Email,
	}, nil
}
