package domain

import (
	"fmt"
)

const (
	// AlreadyExistsErrorType represents the type of an already exists error.
	AlreadyExistsErrorType = "already_exists"
	// DependencyNotFoundErrorType represents the type of a dependency not found error.
	DependencyNotFoundErrorType = "dependency_not_found"
	// FailedDependencyErrorType represents the type of a failed dependency error.
	FailedDependencyErrorType = "failed_dependency"
	// DuplicatedEntityErrorType represents the type of a duplicated entity error.
	DuplicatedEntityErrorType = "duplicated_entity"
	// ForeignKeyViolationErrorType represents the type of a foreign key violation error.
	ForeignKeyViolationErrorType = "foreign_key_violation"
	// NotFoundErrorType represents the type of a not found error.
	NotFoundErrorType = "not_found"
	// ForeignKeyViolationType represents the type of a foreign key violation error.
	ForeignKeyViolationType = "foreign_key_violation"
)

// Error structure that represents a domain error.
type Error struct {
	Type            string `json:"type"`
	Message         string `json:"message"`
	OriginalError   error  `json:"original_error,omitempty"`
	DependencyError []byte `json:"dependency_error,omitempty"`
}

// Error returns the error message from the domain error.
func (e *Error) Error() string {
	return e.Message
}

// NewAlreadyExistsError creates a new domain error for an already existing entity.
func NewAlreadyExistsError(entity, value string, err error) *Error {
	return &Error{
		Type:          AlreadyExistsErrorType,
		Message:       fmt.Sprintf("the %s with value '%s' already exists", entity, value),
		OriginalError: err,
	}
}

// NewDependencyNotFoundError creates a new domain error for a not found dependency.
func NewDependencyNotFoundError(entity, dependency string, value any, err error) *Error {
	return &Error{
		Type:          DependencyNotFoundErrorType,
		Message:       fmt.Sprintf("the %s with value '%v' was not found in the %s", dependency, value, entity),
		OriginalError: err,
	}
}

// NewDuplicatedEntityError creates a new domain error for a duplicated entity.
func NewDuplicatedEntityError(entity, value string, err error) *Error {
	return &Error{
		Type:          DuplicatedEntityErrorType,
		Message:       fmt.Sprintf("the %s with value '%s' is duplicated", entity, value),
		OriginalError: err,
	}
}

// NewForeingKeyMassiveError creates a new domain error for a foreign key violation.
func NewForeingKeyMassiveError(entity string, err error) *Error {
	return &Error{
		Type:          ForeignKeyViolationErrorType,
		Message:       fmt.Sprintf("a record of type %s violates a foreing key", entity),
		OriginalError: err,
	}
}

// NewNotFoundError creates a new domain error for a not found entity.
func NewNotFoundError(entity, value string, err error) *Error {
	return &Error{
		Type:          NotFoundErrorType,
		Message:       fmt.Sprintf("the %s with value '%s' was not found", entity, value),
		OriginalError: err,
	}
}

// NewForeingKeyError creates a new domain error for a foreign key violation.
func NewForeingKeyError(entity, value string, key string, err error) *Error {
	return &Error{
		Type:          ForeignKeyViolationType,
		Message:       fmt.Sprintf("the %s with value '%s' violates the foreing key '%s'", entity, value, key),
		OriginalError: err,
	}
}

// NewFailedDependencyError creates a new domain error for a failed dependency.
func NewFailedDependencyError(dependency string, err error, dependencyErr []byte) *Error {
	return &Error{
		Type:            FailedDependencyErrorType,
		Message:         fmt.Sprintf("the dependency '%s' failed", dependency),
		DependencyError: dependencyErr,
		OriginalError:   err,
	}
}
