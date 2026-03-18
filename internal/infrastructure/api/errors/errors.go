package errors

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/fercho/school-tracking/services/auth/internal/core/auth"
	"github.com/fercho/school-tracking/services/auth/internal/core/domain"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

// applicationType is the content type for JSON responses.
const applicationType = "application/json"

// contentType is the header key for the content type.
const contentType = "Content-Type"

const (
	// InvalidRequestErrorType represents an invalid request error type.
	InvalidRequestErrorType = "invalid_request"
	// UnexpectedErrorType represents an unexpected error type.
	UnexpectedErrorType = "unexpected_error"
)

// statusCodes maps error types to HTTP status codes.
var statusCodes = map[string]int{
	domain.AlreadyExistsErrorType:       http.StatusConflict,
	domain.DependencyNotFoundErrorType:  http.StatusBadRequest,
	domain.DuplicatedEntityErrorType:    http.StatusConflict,
	domain.ForeignKeyViolationErrorType: http.StatusBadRequest,
	domain.NotFoundErrorType:            http.StatusNotFound,
}

// Error structure that represents an API error.
type Error struct {
	Code          int      `json:"code"`
	Type          string   `json:"type"`
	Message       string   `json:"message"`
	Details       []string `json:"details,omitempty"`
	OriginalError string   `json:"-"`
}

// Error returns the error message from the API error.
func (e *Error) Error() string {
	return e.Message
}

// Render renders the API error.
// Method required by the render.Renderer interface.
func (e *Error) Render(w http.ResponseWriter, _ *http.Request) error {
	w.Header().Set(contentType, applicationType)
	w.WriteHeader(e.Code)

	return nil
}

// ParseErrorToResponse parses the error to a JSON response.
func ParseErrorToResponse(_ context.Context, w http.ResponseWriter, r *http.Request, err error) {
	apiErr := MapToAPIError(err)
	_ = render.Render(w, r, apiErr)
}

// newFromDomainError creates an API error from a domain error.
func newFromDomainError(err *domain.Error) *Error {
	code, ok := statusCodes[err.Type]
	if !ok {
		code = http.StatusInternalServerError
	}

	return &Error{
		Code:    code,
		Type:    err.Type,
		Message: err.Message,
	}
}

// MapToAPIError maps an error to an API error.
func MapToAPIError(err error) *Error {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr
	}

	var domainErr *domain.Error
	if errors.As(err, &domainErr) {
		return newFromDomainError(domainErr)
	}

	if errors.Is(err, auth.ErrUserAlreadyExists) {
		return &Error{
			Code:    http.StatusConflict,
			Type:    domain.AlreadyExistsErrorType,
			Message: err.Error(),
		}
	}

	if errors.Is(err, auth.ErrInvalidCredentials) {
		return &Error{
			Code:    http.StatusUnauthorized,
			Type:    InvalidRequestErrorType,
			Message: err.Error(),
		}
	}

	return &Error{
		Code:          http.StatusInternalServerError,
		Type:          UnexpectedErrorType,
		Message:       "there was an unexpected error",
		OriginalError: err.Error(),
	}
}

// NewBodyReadError creates a new body read error.
func NewBodyReadError(err error) *Error {
	return &Error{
		Code:          http.StatusBadRequest,
		Type:          InvalidRequestErrorType,
		Message:       "the request body could not be read",
		OriginalError: err.Error(),
	}
}

// NewBadRequestError creates a new bad request error.
func NewBadRequestError(message string) *Error {
	return &Error{
		Code:    http.StatusBadRequest,
		Type:    InvalidRequestErrorType,
		Message: message,
	}
}

// NewValidatorError creates a new validator error.
func NewValidatorError(err error) *Error {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		var details []string
		for _, vErr := range validationErrors {
			details = append(details, fmt.Sprintf("field '%s' failed validation on '%s' tag", vErr.Field(), vErr.Tag()))
		}
		return &Error{
			Code:          http.StatusBadRequest,
			Type:          InvalidRequestErrorType,
			Message:       "validation failed for one or more fields",
			Details:       details,
			OriginalError: err.Error(),
		}
	}

	return &Error{
		Code:          http.StatusBadRequest,
		Type:          InvalidRequestErrorType,
		Message:       "the request body is not valid",
		OriginalError: err.Error(),
	}
}
