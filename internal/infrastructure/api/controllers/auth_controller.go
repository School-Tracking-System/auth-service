package controllers

import (
	"net/http"

	"github.com/fercho/school-tracking/services/auth/internal/core/domain"
	"github.com/fercho/school-tracking/services/auth/internal/core/ports/services"
	"github.com/fercho/school-tracking/services/auth/internal/infrastructure/api/dtos"
	apierrors "github.com/fercho/school-tracking/services/auth/internal/infrastructure/api/errors"
	"github.com/go-chi/render"
)

// AuthController handles HTTP requests for authentication endpoints.
type AuthController struct {
	authService services.AuthService
}

// NewAuthController creates a new AuthController with the given auth service.
func NewAuthController(authService services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email, password, and role
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dtos.RegisterRequest true "Register Request"
// @Success 201 {object} dtos.UserResponse
// @Failure 400 {object} apierrors.Error
// @Failure 409 {object} apierrors.Error
// @Router /auth/register [post]
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req dtos.RegisterRequest
	if err := bindAndValidateBody(r.Context(), r, &req); err != nil {
		render.Status(r, err.Code)
		render.JSON(w, r, err)
		return
	}

	user, err := c.authService.Register(r.Context(), domain.RegisterParams{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Role:      req.Role,
	})
	if err != nil {
		apiErr := apierrors.MapToAPIError(err)
		render.Status(r, apiErr.Code)
		render.JSON(w, r, apiErr)
		return
	}

	res := dtos.UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		Role:      string(user.Role),
		IsActive:  user.IsActive,
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, res)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dtos.LoginRequest true "Login Request"
// @Success 200 {object} dtos.TokenResponse
// @Failure 400 {object} apierrors.Error
// @Failure 401 {object} apierrors.Error
// @Router /auth/login [post]
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req dtos.LoginRequest
	if err := bindAndValidateBody(r.Context(), r, &req); err != nil {
		render.Status(r, err.Code)
		render.JSON(w, r, err)
		return
	}

	tokens, err := c.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		apiErr := apierrors.MapToAPIError(err)
		render.Status(r, apiErr.Code)
		render.JSON(w, r, apiErr)
		return
	}

	res := dtos.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}

// RefreshToken godoc
// @Summary Refresh JWT Token
// @Description Use refresh token to get a new access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dtos.RefreshRequest true "Refresh Request"
// @Success 200 {object} dtos.TokenResponse
// @Failure 400 {object} apierrors.Error
// @Failure 401 {object} apierrors.Error
// @Router /auth/refresh [post]
func (c *AuthController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dtos.RefreshRequest
	if err := bindAndValidateBody(r.Context(), r, &req); err != nil {
		render.Status(r, err.Code)
		render.JSON(w, r, err)
		return
	}

	tokens, err := c.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		apiErr := apierrors.MapToAPIError(err)
		render.Status(r, apiErr.Code)
		render.JSON(w, r, apiErr)
		return
	}

	res := dtos.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}
