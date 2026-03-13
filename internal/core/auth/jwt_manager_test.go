package auth

import (
	"testing"
	"time"

	"github.com/fercho/school-tracking/services/auth/internal/core/domain"
	"github.com/fercho/school-tracking/services/auth/pkg/env"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestJWTManager() *jwtManager {
	return &jwtManager{
		config: &env.Config{
			JWTSecret: "test-secret-key-for-unit-tests",
		},
	}
}

func TestGenerate(t *testing.T) {
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@school.com",
		Role:  domain.RoleAdmin,
	}

	tests := []struct {
		name    string
		user    *domain.User
		secret  string
		wantErr bool
	}{
		{
			name:   "successful token generation",
			user:   user,
			secret: "test-secret-key-for-unit-tests",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &jwtManager{config: &env.Config{JWTSecret: tc.secret}}

			pair, err := m.Generate(tc.user)

			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, pair.AccessToken)
			assert.NotEmpty(t, pair.RefreshToken)
			assert.NotEqual(t, pair.AccessToken, pair.RefreshToken)
		})
	}
}

func TestValidateAccessToken(t *testing.T) {
	m := newTestJWTManager()
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@school.com",
		Role:  domain.RoleAdmin,
	}

	validPair, err := m.Generate(user)
	require.NoError(t, err)

	expiredToken := generateExpiredAccessToken(t, m.config.JWTSecret, user)

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		wantEmail string
		wantRole  domain.Role
	}{
		{
			name:      "valid access token",
			token:     validPair.AccessToken,
			wantEmail: user.Email,
			wantRole:  user.Role,
		},
		{
			name:    "expired access token",
			token:   expiredToken,
			wantErr: true,
		},
		{
			name:    "malformed token",
			token:   "not.a.valid.jwt",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "token signed with wrong secret",
			token:   generateTokenWithSecret(t, user, "wrong-secret"),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := m.ValidateAccessToken(tc.token)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				require.NoError(t, err)
				assert.Equal(t, user.ID.String(), claims.UserID)
				assert.Equal(t, tc.wantEmail, claims.Email)
				assert.Equal(t, tc.wantRole, claims.Role)
			}
		})
	}
}

func TestValidateRefreshToken(t *testing.T) {
	m := newTestJWTManager()
	user := &domain.User{
		ID:    uuid.New(),
		Email: "test@school.com",
		Role:  domain.RoleAdmin,
	}

	validPair, err := m.Generate(user)
	require.NoError(t, err)

	expiredRefresh := generateExpiredRefreshToken(t, m.config.JWTSecret, user)

	tests := []struct {
		name       string
		token      string
		wantErr    bool
		wantUserID string
	}{
		{
			name:       "valid refresh token",
			token:      validPair.RefreshToken,
			wantUserID: user.ID.String(),
		},
		{
			name:    "expired refresh token",
			token:   expiredRefresh,
			wantErr: true,
		},
		{
			name:    "malformed token",
			token:   "invalid-token",
			wantErr: true,
		},
		{
			name:       "access token used as refresh",
			token:      validPair.AccessToken,
			wantUserID: user.ID.String(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := m.ValidateRefreshToken(tc.token)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantUserID, claims.UserID)
			}
		})
	}
}

// --- Helpers ---

func generateExpiredAccessToken(t *testing.T, secret string, user *domain.User) string {
	t.Helper()
	claims := jwtClaims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Subject:   user.ID.String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return s
}

func generateExpiredRefreshToken(t *testing.T, secret string, user *domain.User) string {
	t.Helper()
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		Subject:   user.ID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return s
}

func generateTokenWithSecret(t *testing.T, user *domain.User, secret string) string {
	t.Helper()
	claims := jwtClaims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return s
}
