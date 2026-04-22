package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fercho/school-tracking/services/auth/internal/core/auth"
	"github.com/fercho/school-tracking/services/auth/internal/core/domain"
	"github.com/fercho/school-tracking/services/auth/internal/core/ports/mocks"
	"github.com/fercho/school-tracking/services/auth/internal/infrastructure/api/dtos"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestController(t *testing.T) (*AuthController, *mocks.MockAuthService) {
	t.Helper()
	mockSvc := mocks.NewMockAuthService(t)
	ctrl := NewAuthController(mockSvc)
	return ctrl, mockSvc
}

func doRequest(handler http.HandlerFunc, method, path string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestRegisterHandler(t *testing.T) {
	registeredUser := &domain.User{
		ID:        uuid.New(),
		Email:     "test@school.com",
		FirstName: "Fernando",
		LastName:  "Garcia",
		Role:      domain.RoleAdmin,
		IsActive:  true,
	}

	tests := []struct {
		name       string
		body       interface{}
		setup      func(svc *mocks.MockAuthService)
		wantStatus int
		wantEmail  string
	}{
		{
			name: "successful register",
			body: dtos.RegisterRequest{
				Email:     "test@school.com",
				Password:  "SecurePass123!",
				FirstName: "Fernando",
				LastName:  "Garcia",
				Role:      domain.RoleAdmin,
			},
			setup: func(svc *mocks.MockAuthService) {
				svc.EXPECT().Register(mock.Anything, mock.AnythingOfType("domain.RegisterParams")).
					Return(registeredUser, nil)
			},
			wantStatus: http.StatusCreated,
			wantEmail:  "test@school.com",
		},
		{
			name: "user already exists",
			body: dtos.RegisterRequest{
				Email:     "test@school.com",
				Password:  "SecurePass123!",
				FirstName: "Fernando",
				LastName:  "Garcia",
				Role:      domain.RoleAdmin,
			},
			setup: func(svc *mocks.MockAuthService) {
				svc.EXPECT().Register(mock.Anything, mock.AnythingOfType("domain.RegisterParams")).
					Return(nil, auth.ErrUserAlreadyExists)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "invalid request body",
			body:       "not-json",
			setup:      func(svc *mocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing email",
			body:       dtos.RegisterRequest{Password: "SecurePass123!", FirstName: "Fernando", LastName: "Garcia", Role: domain.RoleAdmin},
			setup:      func(svc *mocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid email format",
			body:       dtos.RegisterRequest{Email: "not-an-email", Password: "SecurePass123!", FirstName: "Fernando", LastName: "Garcia", Role: domain.RoleAdmin},
			setup:      func(svc *mocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing password",
			body:       dtos.RegisterRequest{Email: "test@school.com", FirstName: "Fernando", LastName: "Garcia", Role: domain.RoleAdmin},
			setup:      func(svc *mocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "password too short",
			body:       dtos.RegisterRequest{Email: "test@school.com", Password: "123", FirstName: "Fernando", LastName: "Garcia", Role: domain.RoleAdmin},
			setup:      func(svc *mocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing first_name",
			body:       dtos.RegisterRequest{Email: "test@school.com", Password: "SecurePass123!", LastName: "Garcia", Role: domain.RoleAdmin},
			setup:      func(svc *mocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing last_name",
			body:       dtos.RegisterRequest{Email: "test@school.com", Password: "SecurePass123!", FirstName: "Fernando", Role: domain.RoleAdmin},
			setup:      func(svc *mocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid role",
			body:       dtos.RegisterRequest{Email: "test@school.com", Password: "SecurePass123!", FirstName: "Fernando", LastName: "Garcia", Role: "superadmin"},
			setup:      func(svc *mocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl, svc := newTestController(t)
			tc.setup(svc)

			rec := doRequest(ctrl.Register, http.MethodPost, "/api/v1/auth/register", tc.body)

			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantEmail != "" {
				var res dtos.UserResponse
				err := json.NewDecoder(rec.Body).Decode(&res)
				require.NoError(t, err)
				assert.Equal(t, tc.wantEmail, res.Email)
				assert.Equal(t, registeredUser.FirstName, res.FirstName)
				assert.True(t, res.IsActive)
			}

			svc.AssertExpectations(t)
		})
	}
}

func TestLoginHandler(t *testing.T) {
	expectedTokens := &domain.TokenPair{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	testUser := &domain.User{
		ID:        uuid.New(),
		Email:     "test@school.com",
		FirstName: "Fernando",
		LastName:  "Garcia",
		Role:      domain.RoleAdmin,
		IsActive:  true,
	}

	tests := []struct {
		name       string
		body       interface{}
		setup      func(svc *mocks.MockAuthService)
		wantStatus int
	}{
		{
			name: "successful login",
			body: dtos.LoginRequest{Email: "test@school.com", Password: "SecurePass123!"},
			setup: func(svc *mocks.MockAuthService) {
				svc.EXPECT().Login(mock.Anything, "test@school.com", "SecurePass123!").
					Return(testUser, expectedTokens, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid credentials",
			body: dtos.LoginRequest{Email: "test@school.com", Password: "wrong"},
			setup: func(svc *mocks.MockAuthService) {
				svc.EXPECT().Login(mock.Anything, "test@school.com", "wrong").
					Return(nil, nil, auth.ErrInvalidCredentials)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid request body",
			body:       "bad-json",
			setup:      func(svc *mocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing email",
			body:       dtos.LoginRequest{Password: "SecurePass123!"},
			setup:      func(svc *mocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing password",
			body:       dtos.LoginRequest{Email: "test@school.com"},
			setup:      func(svc *mocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid email format",
			body:       dtos.LoginRequest{Email: "not-an-email", Password: "SecurePass123!"},
			setup:      func(svc *mocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl, svc := newTestController(t)
			tc.setup(svc)

			rec := doRequest(ctrl.Login, http.MethodPost, "/api/v1/auth/login", tc.body)

			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantStatus == http.StatusOK {
				var res dtos.AuthResponse
				err := json.NewDecoder(rec.Body).Decode(&res)
				require.NoError(t, err)
				assert.Equal(t, testUser.Email, res.User.Email)
				assert.Equal(t, expectedTokens.AccessToken, res.Tokens.AccessToken)
				assert.Equal(t, expectedTokens.RefreshToken, res.Tokens.RefreshToken)
			}

			svc.AssertExpectations(t)
		})
	}
}

func TestRefreshTokenHandler(t *testing.T) {
	expectedTokens := &domain.TokenPair{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}

	tests := []struct {
		name       string
		body       interface{}
		setup      func(svc *mocks.MockAuthService)
		wantStatus int
	}{
		{
			name: "successful refresh",
			body: dtos.RefreshRequest{RefreshToken: "valid-refresh-token"},
			setup: func(svc *mocks.MockAuthService) {
				svc.EXPECT().RefreshToken(mock.Anything, "valid-refresh-token").
					Return(expectedTokens, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid refresh token",
			body: dtos.RefreshRequest{RefreshToken: "invalid-token"},
			setup: func(svc *mocks.MockAuthService) {
				svc.EXPECT().RefreshToken(mock.Anything, "invalid-token").
					Return(nil, errors.New("invalid token"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid request body",
			body:       "bad-json",
			setup:      func(svc *mocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing refresh_token",
			body:       dtos.RefreshRequest{},
			setup:      func(svc *mocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl, svc := newTestController(t)
			tc.setup(svc)

			rec := doRequest(ctrl.RefreshToken, http.MethodPost, "/api/v1/auth/refresh", tc.body)

			assert.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantStatus == http.StatusOK {
				var res dtos.TokenResponse
				err := json.NewDecoder(rec.Body).Decode(&res)
				require.NoError(t, err)
				assert.Equal(t, expectedTokens.AccessToken, res.AccessToken)
				assert.Equal(t, expectedTokens.RefreshToken, res.RefreshToken)
			}

			svc.AssertExpectations(t)
		})
	}
}
