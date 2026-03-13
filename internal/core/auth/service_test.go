package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/fercho/school-tracking/services/auth/internal/core/domain"
	"github.com/fercho/school-tracking/services/auth/internal/core/ports/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func newTestService(t *testing.T) (*authService, *mocks.MockUserRepository, *mocks.MockJWTManager) {
	t.Helper()
	repo := mocks.NewMockUserRepository(t)
	jwt := mocks.NewMockJWTManager(t)
	svc := &authService{userRepo: repo, jwtManager: jwt}
	return svc, repo, jwt
}

func TestRegister(t *testing.T) {
	validParams := domain.RegisterParams{
		Email:     "test@school.com",
		Password:  "SecurePass123!",
		FirstName: "Fernando",
		LastName:  "Garcia",
		Role:      domain.RoleAdmin,
	}

	tests := []struct {
		name      string
		params    domain.RegisterParams
		setup     func(repo *mocks.MockUserRepository, jwt *mocks.MockJWTManager)
		wantErr   error
		wantEmail string
	}{
		{
			name:   "successful registration",
			params: validParams,
			setup: func(repo *mocks.MockUserRepository, jwt *mocks.MockJWTManager) {
				repo.EXPECT().GetByEmail(mock.Anything, validParams.Email).
					Return(nil, errors.New("not found"))
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.User")).
					Return(nil)
			},
			wantEmail: validParams.Email,
		},
		{
			name:   "user already exists",
			params: validParams,
			setup: func(repo *mocks.MockUserRepository, jwt *mocks.MockJWTManager) {
				repo.EXPECT().GetByEmail(mock.Anything, validParams.Email).
					Return(&domain.User{Email: validParams.Email}, nil)
			},
			wantErr: ErrUserAlreadyExists,
		},
		{
			name:   "repository create error",
			params: validParams,
			setup: func(repo *mocks.MockUserRepository, jwt *mocks.MockJWTManager) {
				repo.EXPECT().GetByEmail(mock.Anything, validParams.Email).
					Return(nil, errors.New("not found"))
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.User")).
					Return(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, jwt := newTestService(t)
			tc.setup(repo, jwt)

			user, err := svc.Register(context.Background(), tc.params)

			if tc.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tc.wantEmail, user.Email)
				assert.Equal(t, tc.params.FirstName, user.FirstName)
				assert.Equal(t, tc.params.LastName, user.LastName)
				assert.Equal(t, tc.params.Role, user.Role)
				assert.True(t, user.IsActive)
				assert.NotEqual(t, uuid.Nil, user.ID)
				assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(tc.params.Password)))
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestLogin(t *testing.T) {
	hashedPw, _ := bcrypt.GenerateFromPassword([]byte("SecurePass123!"), bcrypt.DefaultCost)
	existingUser := &domain.User{
		ID:           uuid.New(),
		Email:        "test@school.com",
		PasswordHash: string(hashedPw),
		Role:         domain.RoleAdmin,
	}
	expectedTokens := &domain.TokenPair{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	tests := []struct {
		name     string
		email    string
		password string
		setup    func(repo *mocks.MockUserRepository, jwt *mocks.MockJWTManager)
		wantErr  error
	}{
		{
			name:     "successful login",
			email:    "test@school.com",
			password: "SecurePass123!",
			setup: func(repo *mocks.MockUserRepository, jwt *mocks.MockJWTManager) {
				repo.EXPECT().GetByEmail(mock.Anything, "test@school.com").
					Return(existingUser, nil)
				jwt.EXPECT().Generate(existingUser).
					Return(expectedTokens, nil)
			},
		},
		{
			name:     "user not found",
			email:    "unknown@school.com",
			password: "SecurePass123!",
			setup: func(repo *mocks.MockUserRepository, jwt *mocks.MockJWTManager) {
				repo.EXPECT().GetByEmail(mock.Anything, "unknown@school.com").
					Return(nil, errors.New("not found"))
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name:     "wrong password",
			email:    "test@school.com",
			password: "WrongPassword!",
			setup: func(repo *mocks.MockUserRepository, jwt *mocks.MockJWTManager) {
				repo.EXPECT().GetByEmail(mock.Anything, "test@school.com").
					Return(existingUser, nil)
			},
			wantErr: ErrInvalidCredentials,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, jwt := newTestService(t)
			tc.setup(repo, jwt)

			tokens, err := svc.Login(context.Background(), tc.email, tc.password)

			if tc.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, tokens)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedTokens, tokens)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestValidateToken(t *testing.T) {
	expectedClaims := &domain.Claims{
		UserID: uuid.New().String(),
		Email:  "test@school.com",
		Role:   domain.RoleAdmin,
	}

	tests := []struct {
		name    string
		token   string
		setup   func(repo *mocks.MockUserRepository, jwt *mocks.MockJWTManager)
		wantErr bool
	}{
		{
			name:  "valid token",
			token: "valid-access-token",
			setup: func(repo *mocks.MockUserRepository, jwt *mocks.MockJWTManager) {
				jwt.EXPECT().ValidateAccessToken("valid-access-token").
					Return(expectedClaims, nil)
			},
		},
		{
			name:  "invalid token",
			token: "invalid-token",
			setup: func(repo *mocks.MockUserRepository, jwt *mocks.MockJWTManager) {
				jwt.EXPECT().ValidateAccessToken("invalid-token").
					Return(nil, ErrInvalidToken)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, jwt := newTestService(t)
			tc.setup(repo, jwt)

			claims, err := svc.ValidateToken(context.Background(), tc.token)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedClaims, claims)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestRefreshToken(t *testing.T) {
	userID := uuid.New()
	existingUser := &domain.User{
		ID:    userID,
		Email: "test@school.com",
		Role:  domain.RoleAdmin,
	}
	expectedTokens := &domain.TokenPair{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}

	tests := []struct {
		name    string
		token   string
		setup   func(repo *mocks.MockUserRepository, jwt *mocks.MockJWTManager)
		wantErr bool
	}{
		{
			name:  "successful refresh",
			token: "valid-refresh-token",
			setup: func(repo *mocks.MockUserRepository, jwt *mocks.MockJWTManager) {
				jwt.EXPECT().ValidateRefreshToken("valid-refresh-token").
					Return(&domain.Claims{UserID: userID.String()}, nil)
				repo.EXPECT().GetByID(mock.Anything, userID).
					Return(existingUser, nil)
				jwt.EXPECT().Generate(existingUser).
					Return(expectedTokens, nil)
			},
		},
		{
			name:  "invalid refresh token",
			token: "invalid-refresh-token",
			setup: func(repo *mocks.MockUserRepository, jwt *mocks.MockJWTManager) {
				jwt.EXPECT().ValidateRefreshToken("invalid-refresh-token").
					Return(nil, ErrInvalidToken)
			},
			wantErr: true,
		},
		{
			name:  "user not found on refresh",
			token: "valid-refresh-token",
			setup: func(repo *mocks.MockUserRepository, jwt *mocks.MockJWTManager) {
				jwt.EXPECT().ValidateRefreshToken("valid-refresh-token").
					Return(&domain.Claims{UserID: userID.String()}, nil)
				repo.EXPECT().GetByID(mock.Anything, userID).
					Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, jwt := newTestService(t)
			tc.setup(repo, jwt)

			tokens, err := svc.RefreshToken(context.Background(), tc.token)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, tokens)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedTokens, tokens)
			}

			repo.AssertExpectations(t)
		})
	}
}
