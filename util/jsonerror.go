package util

import (
	"fmt"
	"net/http"
)

// PfsError represents the standard error response in pfs server.
type PfsError struct {
	ErrCode string `json:"errcode"`
	Err     string `json:"error"`
}

// Unknown is an unexpected error
func Unknown(msg string) *PfsError {
	return &PfsError{"M_UNKNOWN", msg}
}

// Error system error
func (e *PfsError) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrCode, e.Err)
}

// InternalServerError returns a 500 Internal Server Error in a pfs-compliant format.
func InternalServerError() JSONResponse {
	return JSONResponse{
		Code: http.StatusInternalServerError,
		JSON: Unknown("Internal Server Error"),
	}
}

// OkJSON is a result of no-error
func OkJSON(msg string) *PfsError {
	return &PfsError{"M_OK", msg}
}

// Forbidden is an error when the client tries to access a resource
// they are not allowed to access.
func Forbidden(msg string) *PfsError {
	return &PfsError{"M_FORBIDDEN", msg}
}

// BadJSON is an error when the client supplies malformed JSON.
func BadJSON(msg string) *PfsError {
	return &PfsError{"M_BAD_JSON", msg}
}

// NotJSON is an error when the client supplies something that is not JSON
// to a JSON endpoint.
func NotJSON(msg string) *PfsError {
	return &PfsError{"M_NOT_JSON", msg}
}

// NotFound is an error when the client tries to access an unknown resource.
func NotFound(msg string) *PfsError {
	return &PfsError{"M_NOT_FOUND", msg}
}

// MissingArgument is an error when the client tries to access a resource
// without providing an argument that is required.
func MissingArgument(msg string) *PfsError {
	return &PfsError{"M_MISSING_ARGUMENT", msg}
}

// InvalidArgumentValue is an error when the client tries to provide an
// invalid value for a valid argument
func InvalidArgumentValue(msg string) *PfsError {
	return &PfsError{"M_INVALID_ARGUMENT_VALUE", msg}
}

// MissingToken is an error when the client tries to access a resource which
// requires authentication without supplying credentials.
func MissingToken(msg string) *PfsError {
	return &PfsError{"M_MISSING_TOKEN", msg}
}

// UnknownToken is an error when the client tries to access a resource which
// requires authentication and supplies an unrecognized token
func UnknownToken(msg string) *PfsError {
	return &PfsError{"M_UNKNOWN_TOKEN", msg}
}


// ASExclusive is an error returned when an application service tries to
// register an username that is outside of its registered namespace, or if a
// user attempts to register a username or room alias within an exclusive
// namespace.
func ASExclusive(msg string) *PfsError {
	return &PfsError{"M_EXCLUSIVE", msg}
}


// LimitExceededError is a rate-limiting error.
type LimitExceededError struct {
	PfsError
	RetryAfterMS int64 `json:"retry_after_ms,omitempty"`
}

// LimitExceeded is an error when the client tries to send events too quickly.
func LimitExceeded(msg string, retryAfterMS int64) *LimitExceededError {
	return &LimitExceededError{
		PfsError:  PfsError{"M_LIMIT_EXCEEDED", msg},
		RetryAfterMS: retryAfterMS,
	}
}

// NotTrusted is an error which is returned when the client asks the server to
// proxy a request (e.g. 3PID association) to a server that isn't trusted
func NotTrusted(serverName string) *PfsError {
	return &PfsError{
		ErrCode: "M_SERVER_NOT_TRUSTED",
		Err:     fmt.Sprintf("Untrusted server '%s'", serverName),
	}
}