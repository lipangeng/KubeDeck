package errors

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// ErrorCode represents application error codes
type ErrorCode string

// Standard error codes
const (
	ErrUnknown        ErrorCode = "UNKNOWN"
	ErrValidation     ErrorCode = "VALIDATION_ERROR"
	ErrAuthentication ErrorCode = "AUTHENTICATION_ERROR"
	ErrAuthorization  ErrorCode = "AUTHORIZATION_ERROR"
	ErrNotFound       ErrorCode = "NOT_FOUND"
	ErrConflict       ErrorCode = "CONFLICT"
	ErrRateLimit      ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrInternal       ErrorCode = "INTERNAL_ERROR"
	ErrBadRequest     ErrorCode = "BAD_REQUEST"
	ErrTimeout        ErrorCode = "TIMEOUT"
	ErrUnavailable    ErrorCode = "SERVICE_UNAVAILABLE"
	ErrKubernetes     ErrorCode = "KUBERNETES_ERROR"
	ErrDatabase       ErrorCode = "DATABASE_ERROR"
	ErrEncryption     ErrorCode = "ENCRYPTION_ERROR"
)

// AppError represents an application error with context
type AppError struct {
	Code       ErrorCode         `json:"code"`
	Message    string            `json:"message"`
	Details    map[string]string `json:"details,omitempty"`
	Cause      error             `json:"-"`
	HTTPStatus int               `json:"-"`
	Timestamp  time.Time         `json:"timestamp"`
	TraceID    string            `json:"trace_id,omitempty"`
	File       string            `json:"-"`
	Line       int               `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Builder creates AppError with fluent interface
type Builder struct {
	code       ErrorCode
	message    string
	details    map[string]string
	cause      error
	httpStatus int
	traceID    string
}

// New creates a new error builder
func New(code ErrorCode, message string) *Builder {
	return &Builder{
		code:    code,
		message: message,
		details: make(map[string]string),
	}
}

// WithDetails adds details to the error
func (b *Builder) WithDetails(key, value string) *Builder {
	b.details[key] = value
	return b
}

// WithCause sets the underlying error
func (b *Builder) WithCause(err error) *Builder {
	b.cause = err
	return b
}

// WithHTTPStatus sets the HTTP status code
func (b *Builder) WithHTTPStatus(status int) *Builder {
	b.httpStatus = status
	return b
}

// WithTraceID sets the trace ID
func (b *Builder) WithTraceID(traceID string) *Builder {
	b.traceID = traceID
	return b
}

// Build creates the AppError
func (b *Builder) Build() *AppError {
	_, file, line, _ := runtime.Caller(2)
	shortFile := file
	if idx := strings.LastIndex(file, "/"); idx != -1 {
		shortFile = file[idx+1:]
	}

	return &AppError{
		Code:       b.code,
		Message:    b.message,
		Details:    b.details,
		Cause:      b.cause,
		HTTPStatus: b.httpStatus,
		Timestamp:  time.Now(),
		TraceID:    b.traceID,
		File:       shortFile,
		Line:       line,
	}
}

// Predefined error constructors

// NewValidation creates a validation error
func NewValidation(message string, field string) *AppError {
	return New(ErrValidation, message).
		WithDetails("field", field).
		WithHTTPStatus(http.StatusBadRequest).
		Build()
}

// NewAuth creates an authentication error
func NewAuth(message string) *AppError {
	return New(ErrAuthentication, message).
		WithHTTPStatus(http.StatusUnauthorized).
		Build()
}

// NewForbidden creates an authorization error
func NewForbidden(message string) *AppError {
	return New(ErrAuthorization, message).
		WithHTTPStatus(http.StatusForbidden).
		Build()
}

// NewNotFound creates a not found error
func NewNotFound(resource, id string) *AppError {
	return New(ErrNotFound, fmt.Sprintf("%s not found: %s", resource, id)).
		WithDetails("resource", resource).
		WithDetails("id", id).
		WithHTTPStatus(http.StatusNotFound).
		Build()
}

// NewConflict creates a conflict error
func NewConflict(resource, field, value string) *AppError {
	return New(ErrConflict, fmt.Sprintf("%s conflict: %s=%s", resource, field, value)).
		WithDetails("resource", resource).
		WithDetails("field", field).
		WithDetails("value", value).
		WithHTTPStatus(http.StatusConflict).
		Build()
}

// NewRateLimit creates a rate limit error
func NewRateLimit(retryAfter int) *AppError {
	return New(ErrRateLimit, "Rate limit exceeded").
		WithDetails("retry_after", fmt.Sprintf("%d", retryAfter)).
		WithHTTPStatus(http.StatusTooManyRequests).
		Build()
}

// NewInternal creates an internal error
func NewInternal(message string, cause error) *AppError {
	return New(ErrInternal, message).
		WithCause(cause).
		WithHTTPStatus(http.StatusInternalServerError).
		Build()
}

// NewBadRequest creates a bad request error
func NewBadRequest(message string) *AppError {
	return New(ErrBadRequest, message).
		WithHTTPStatus(http.StatusBadRequest).
		Build()
}

// NewTimeout creates a timeout error
func NewTimeout(operation string) *AppError {
	return New(ErrTimeout, fmt.Sprintf("Operation timed out: %s", operation)).
		WithHTTPStatus(http.StatusRequestTimeout).
		Build()
}

// NewUnavailable creates a service unavailable error
func NewUnavailable(service string) *AppError {
	return New(ErrUnavailable, fmt.Sprintf("Service unavailable: %s", service)).
		WithHTTPStatus(http.StatusServiceUnavailable).
		Build()
}

// NewKubernetes creates a Kubernetes error
func NewKubernetes(message string, cause error) *AppError {
	return New(ErrKubernetes, message).
		WithCause(cause).
		WithHTTPStatus(http.StatusInternalServerError).
		Build()
}

// NewDatabase creates a database error
func NewDatabase(message string, cause error) *AppError {
	return New(ErrDatabase, message).
		WithCause(cause).
		WithHTTPStatus(http.StatusInternalServerError).
		Build()
}

// NewEncryption creates an encryption error
func NewEncryption(message string, cause error) *AppError {
	return New(ErrEncryption, message).
		WithCause(cause).
		WithHTTPStatus(http.StatusInternalServerError).
		Build()
}

// Utility functions

// Is checks if error is of a specific type
func Is(err error, target error) bool {
	return errors.Is(err, target)
}

// As attempts to extract an AppError
func As(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// GetHTTPStatus returns HTTP status for an error
func GetHTTPStatus(err error) int {
	if appErr, ok := As(err); ok {
		if appErr.HTTPStatus != 0 {
			return appErr.HTTPStatus
		}
	}

	// Default mappings based on error code
	if appErr, ok := As(err); ok {
		switch appErr.Code {
		case ErrValidation, ErrBadRequest:
			return http.StatusBadRequest
		case ErrAuthentication:
			return http.StatusUnauthorized
		case ErrAuthorization:
			return http.StatusForbidden
		case ErrNotFound:
			return http.StatusNotFound
		case ErrConflict:
			return http.StatusConflict
		case ErrRateLimit:
			return http.StatusTooManyRequests
		}
	}

	return http.StatusInternalServerError
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error     ErrorCode         `json:"error"`
	Message   string            `json:"message"`
	Details   map[string]string `json:"details,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	TraceID   string            `json:"trace_id,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

// ToResponse converts AppError to API response
func (e *AppError) ToResponse(requestID string) ErrorResponse {
	return ErrorResponse{
		Error:     e.Code,
		Message:   e.Message,
		Details:   e.Details,
		Timestamp: e.Timestamp,
		TraceID:   e.TraceID,
		RequestID: requestID,
	}
}
