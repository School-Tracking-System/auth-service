package auth

import (
	"context"
	"errors"

	"github.com/fercho/school-tracking/services/auth/internal/core/domain"
	"github.com/fercho/school-tracking/services/auth/internal/core/ports/repositories"
	"github.com/fercho/school-tracking/services/auth/internal/core/ports/services"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type authService struct {
	userRepo   repositories.UserRepository
	jwtManager services.JWTManager
}

// NewAuthService creates a new AuthService with the given repository and JWT manager.
func NewAuthService(userRepo repositories.UserRepository, jwtManager services.JWTManager) services.AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register creates a new user after validating uniqueness and hashing the password.
func (s *authService) Register(ctx context.Context, params domain.RegisterParams) (*domain.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, params.Email)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           uuid.New(),
		Email:        params.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    params.FirstName,
		LastName:     params.LastName,
		Phone:        params.Phone,
		Role:         params.Role,
		IsActive:     true,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user by email and password, returning a JWT token pair.
func (s *authService) Login(ctx context.Context, email, password string) (*domain.TokenPair, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.jwtManager.Generate(user)
}

// ValidateToken verifies an access token and returns its decoded claims.
func (s *authService) ValidateToken(ctx context.Context, token string) (*domain.Claims, error) {
	return s.jwtManager.ValidateAccessToken(token)
}

// RefreshToken validates a refresh token and issues a new token pair for the user.
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	claims, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.jwtManager.Generate(user)
}
