package sonzai

import "fmt"

// SonzaiError is the base error type for all SDK errors.
type SonzaiError struct {
	StatusCode int
	Message    string
}

func (e *SonzaiError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("[%d] %s", e.StatusCode, e.Message)
	}
	return e.Message
}

// AuthenticationError is returned when the API key is invalid or missing.
type AuthenticationError struct{ SonzaiError }

// NotFoundError is returned when the requested resource is not found.
type NotFoundError struct{ SonzaiError }

// BadRequestError is returned when the request is invalid.
type BadRequestError struct{ SonzaiError }

// PermissionDeniedError is returned when the API key lacks permission.
type PermissionDeniedError struct{ SonzaiError }

// RateLimitError is returned when rate limit is exceeded.
// RetryAfter holds the number of seconds to wait before retrying, if the
// server returned a Retry-After header.
type RateLimitError struct {
	SonzaiError
	RetryAfter *float64
}

// InternalServerError is returned when the server returns a 5xx error.
type InternalServerError struct{ SonzaiError }

func newErrorForStatus(statusCode int, message string) error {
	base := SonzaiError{StatusCode: statusCode, Message: message}
	switch statusCode {
	case 401:
		return &AuthenticationError{base}
	case 403:
		return &PermissionDeniedError{base}
	case 404:
		return &NotFoundError{base}
	case 400:
		return &BadRequestError{base}
	case 429:
		return &RateLimitError{SonzaiError: base}
	default:
		if statusCode >= 500 {
			return &InternalServerError{base}
		}
		return &base
	}
}
