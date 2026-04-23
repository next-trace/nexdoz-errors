package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	NotFoundErrorType            = "NotFoundError"
	InternalServerErrorType      = "InternalServerError"
	BadRequestErrorType          = "BadRequestError"
	UnauthorizedErrorType        = "UnauthorizedError"
	ForbiddenErrorType           = "ForbiddenError"
	ConflictErrorType            = "ConflictError"
	MethodNotAllowedErrorType    = "MethodNotAllowedError"
	RequestTimeoutErrorType      = "RequestTimeoutError"
	UnprocessableEntityErrorType = "UnprocessableEntityError"
	TooManyRequestsErrorType     = "TooManyRequestsError"
	GenericErrorType             = "GenericError"
)

// ErrorOption configures an ApiError during construction.
type ErrorOption func(*ApiError)

// ApiErrors is the public contract every API error in the diabuddy platform
// implements. Callers depending on this interface get access to structured
// metadata (type, code), stable string formatting, HTTP mapping, JSON
// (un)marshaling, and the wrapped internal error for errors.Is / errors.As
// chain walking.
type ApiErrors interface {
	error
	json.Marshaler
	json.Unmarshaler
	Type() string
	Code() int
	HTTPError() (int, string)
	InternalError() error
	Unwrap() error
}

// ApiError represents a structured error for the API.
type ApiError struct {
	ErrorType  string `json:"error_type"`
	Message    string `json:"message"`
	ErrorCode  int    `json:"error_code"`
	InnerError error  `json:"-"`
}

// compile-time check: ApiError satisfies the ApiErrors interface.
var _ ApiErrors = (*ApiError)(nil)

// Type returns the ApiError type name.
func (e *ApiError) Type() string {
	return e.ErrorType
}

// Code returns the ApiError HTTP status code.
func (e *ApiError) Code() int {
	return e.ErrorCode
}

// Error implements the error interface for ApiError.
func (e *ApiError) Error() string {
	return fmt.Sprintf("Error %d: %s", e.ErrorCode, e.Message)
}

// HTTPError returns the status code and user-facing message pair suitable for
// direct HTTP response writing.
func (e *ApiError) HTTPError() (int, string) {
	return e.ErrorCode, e.Message
}

// InternalError returns the wrapped internal cause, or nil if none was set.
func (e *ApiError) InternalError() error {
	return e.InnerError
}

// Unwrap exposes the wrapped internal cause so errors.Is and errors.As can
// walk the error chain.
func (e *ApiError) Unwrap() error {
	return e.InnerError
}

// MarshalJSON customises the JSON serialisation for ApiError by inlining the
// internal error message as a top-level "internal_error" field when present.
func (e *ApiError) MarshalJSON() ([]byte, error) {
	type Alias ApiError // Alias to avoid recursion.
	var internalErrorStr string
	if e.InnerError != nil {
		internalErrorStr = e.InnerError.Error()
	}

	return json.Marshal(&struct {
		InternalError string `json:"internal_error,omitempty"`
		*Alias
	}{
		InternalError: internalErrorStr,
		Alias:         (*Alias)(e),
	})
}

// UnmarshalJSON customises the JSON deserialisation for ApiError. If an
// "internal_error" field is present, it is rewrapped as a plain error via
// errors.New so any percent signs in the payload are treated literally.
func (e *ApiError) UnmarshalJSON(data []byte) error {
	type Alias ApiError
	aux := &struct {
		InternalError *string `json:"internal_error,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.InternalError != nil {
		e.InnerError = errors.New(*aux.InternalError)
	}
	return nil
}

// ErrorType represents an error type configuration in the registry.
type ErrorType struct {
	ErrorCode int
	Message   string
}

// ErrorRegistry is the default mapping of error type names to their HTTP code
// and canonical message. Mutate only at init time; see RegisterErrorType.
var ErrorRegistry = map[string]ErrorType{
	NotFoundErrorType:            {http.StatusNotFound, "Resource not found"},
	InternalServerErrorType:      {http.StatusInternalServerError, "Internal server error"},
	BadRequestErrorType:          {http.StatusBadRequest, "Bad request"},
	UnauthorizedErrorType:        {http.StatusUnauthorized, "Unauthorized access"},
	ForbiddenErrorType:           {http.StatusForbidden, "Forbidden"},
	ConflictErrorType:            {http.StatusConflict, "Conflict occurred"},
	MethodNotAllowedErrorType:    {http.StatusMethodNotAllowed, "Method not allowed"},
	RequestTimeoutErrorType:      {http.StatusRequestTimeout, "Request timed out"},
	UnprocessableEntityErrorType: {http.StatusUnprocessableEntity, "Unprocessable entity"},
	TooManyRequestsErrorType:     {http.StatusTooManyRequests, "Too many requests"},
}

// NewApiError creates a new ApiError. If errorType is registered in
// ErrorRegistry, its code is applied; otherwise the result falls back to
// GenericErrorType + 500. Functional options run last and can override any
// field.
func NewApiError(errorType string, userMessage string, options ...ErrorOption) *ApiError {
	apiError := &ApiError{
		ErrorType: GenericErrorType,
		Message:   userMessage,
		ErrorCode: http.StatusInternalServerError,
	}
	if entry, exists := ErrorRegistry[errorType]; exists {
		apiError.ErrorType = errorType
		apiError.ErrorCode = entry.ErrorCode
	}
	for _, option := range options {
		option(apiError)
	}
	return apiError
}

// RegisterErrorType adds a new error type to ErrorRegistry.
//
// NOT safe for concurrent use. Call it only during package initialisation
// (for example, from init() functions or program startup), before any
// goroutines read from ErrorRegistry.
func RegisterErrorType(name string, errorCode int, message string) {
	ErrorRegistry[name] = ErrorType{errorCode, message}
}

// WithInternalError attaches a wrapped internal cause to the ApiError. The
// cause is accessible via InternalError, Unwrap, errors.Is and errors.As.
func WithInternalError(err error) ErrorOption {
	return func(ae *ApiError) {
		ae.InnerError = err
	}
}
