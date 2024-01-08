package app

import (
	"errors"
	"fmt"
	"net/http"
)

// The default error to use if an error is not defined when creating a WrappedSafeError with the Wrap
// function.
var DefaultError = errors.New("undefined error")

// The default safe message to use if a safe message is not defined when creating a WrappedSafeError
// with the Wrap function.
var DefaultSafeMessage = "Something went wrong. Please try again later."

// The default status code to use if a status code is not defined when creating a WrappedSafeError
// with the Wrap function.
var DefaultStatusCode = http.StatusInternalServerError

// SafeError is the interface that wraps the Safe func.
//
// When an error is returned to a HTTP handler, if the error can be converted to a SafeError using
// errors.As, the handler should write the message and status code returned by Safe.
//
// If an error cannot be converted to a SafeError, the cause of the error should be hidden.
type SafeError interface {
	// Safe returns a user friendly error message and the HTTP status code.
	Safe() (string, int)
}

// WrappedSafeError satisfies the error and SafeError interface.
//
// If a particular error needs to display safe informative data to the user, return a WrappedSafeError
// when returning an error.
//
// WrappedSafeError should only be created using the wrapper functions.
type WrappedSafeError struct {
	// The underlying error. This is the error that will be returned by the error interface Error func.
	// If an error is not defined it will default to DefaultError.
	err error

	// The user friendly error message. If a safe message is not defined it will default to
	// DefaultSafeMessage.
	safeMessage string

	// The HTTP status code that should be used. If a HTTP status code is not defined it will default
	// to DefaultStatusCode.
	statusCode int
}

// WrapParams is the paramters used when wrapping an error with the Wrap func.
type WrapParams struct {
	// The error to be wrapped. This will be set as the underylying error in the WrappedSafeError.
	Err error

	// The user friendly error message that will be set in the WrappedSafeError. This value will
	// be returned by the Safe method.
	SafeMessage string

	// The HTTP status code that will be set in the WrappedSafeError. This value will be returned
	// by the Safe method.
	StatusCode int
}

// Wrap will wrap an error in a WrappedSafeError.
//
// If Err, SafeMessage, or StatusCode is not set, WrappedSafeError will be set with DefaultError,
// DefaultSafeMessage, or DefaultStatusCode.
func Wrap(w WrapParams) *WrappedSafeError {
	if w.Err == nil {
		w.Err = DefaultError
	}

	if w.SafeMessage == "" {
		w.SafeMessage = DefaultSafeMessage
	}

	if w.StatusCode == 0 {
		w.StatusCode = DefaultStatusCode
	}

	return &WrappedSafeError{
		err:         w.Err,
		safeMessage: w.SafeMessage,
		statusCode:  w.StatusCode,
	}
}

// Error is the function that satisfies the error interface.
func (e *WrappedSafeError) Error() string {
	return e.err.Error()
}

// Unwrap is the function that will be used by errors.Is to compare the wrapped error.
func (e *WrappedSafeError) Unwrap() error {
	return e.err
}

// Safe is the function that satisfies the SafeError interface.
//
// Safe returns the user friendly error message and the HTTP status code.
func (e *WrappedSafeError) Safe() (string, int) {
	return e.safeMessage, e.statusCode
}

// WriteJSONError writes a error message body to w. If the error can be converted to a SafeError, the message
// and status code returned by the Safe function will be displayed to the end user. If the error cannot be
// converted, the cause of the error will remain hidden from the end user. In this case, a generic message
// will be displayed along with a 500 Internal Server Error status code.
func WriteJSONError(w http.ResponseWriter, err error) {
	var (
		safeError  SafeError
		message    string
		statusCode int
	)

	if errors.As(err, &safeError) {
		message, statusCode = safeError.Safe()
	} else {
		message = DefaultSafeMessage
		statusCode = DefaultStatusCode
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(JSONError(message, statusCode))
}

// JSONError creates a JSON error response with a "error" and "status_code" field. The "error" field should
// be a safe message that can be displayed to an end user.
func JSONError(err string, statusCode int) []byte {
	str := fmt.Sprintf(`{"error": "%s", "status_code": %d}`, err, statusCode)
	return []byte(str)
}
