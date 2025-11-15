package domain

import "errors"

// ErrorCode categorizes high level error conditions for uniform handling.
type ErrorCode string

const (
	ErrorCodeValidation ErrorCode = "validation_error"
	ErrorCodeNotFound   ErrorCode = "not_found"
	ErrorCodeInternal   ErrorCode = "internal_error"
)

// Error represents a domain level error with semantic information.
type Error struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *Error) Error() string {
	return e.Message
}

// Unwrap allows errors.Is/As to traverse the wrapped error.
func (e *Error) Unwrap() error {
	return e.Err
}

// NewValidationError creates a validation error with an optional wrapped error.
func NewValidationError(message string, err error) *Error {
	return &Error{
		Code:    ErrorCodeValidation,
		Message: message,
		Err:     err,
	}
}

// NewNotFoundError returns a not found error.
func NewNotFoundError(message string) *Error {
	return &Error{
		Code:    ErrorCodeNotFound,
		Message: message,
		Err:     errors.New(message),
	}
}

// NewInternalError wraps unexpected failures.
func NewInternalError(message string, err error) *Error {
	return &Error{
		Code:    ErrorCodeInternal,
		Message: message,
		Err:     err,
	}
}
