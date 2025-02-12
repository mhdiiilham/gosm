package entity

import "fmt"

// GosmErrorType defines a type for categorizing errors within the application.
type GosmErrorType string

// Predefined error types to categorize different error scenarios.
var (
	GosmErrorTypeBadRequest GosmErrorType = "4" // Represents client-side errors (e.g., validation failures)
	GosmErrorTypeUnknown    GosmErrorType = "5" // Represents unexpected or internal server errors
)

// GosmError represents a structured application error.
// It includes an error type, code, message, and the underlying source error.
type GosmError struct {
	Type    GosmErrorType `json:"type"`    // Category of the error
	Code    string        `json:"code"`    // Unique error code for identification
	Message string        `json:"message"` // Human-readable error message
	Source  error         `json:"error"`   // Underlying error (if any)
}

// Error implements the error interface, returning a formatted string representation of the error.
func (e GosmError) Error() string {
	return fmt.Sprintf("type:%s | code:%s | message:%s", e.Type, e.Code, e.Message)
}

// UnknownError wraps an unexpected error in a GosmError with an "INTERNAL_SERVER_ERROR" code.
// It provides a standardized way to represent internal server errors.
func UnknownError(err error) error {
	return GosmError{
		Type:    GosmErrorTypeUnknown,
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "Unexpected error occured",
		Source:  err,
	}
}

// NewBadRequestError creates a new instance of GosmError representing a bad request error.
// It is used when the client sends an invalid request.
func NewBadRequestError(code string, message string) error {
	return GosmError{
		Type:    GosmErrorTypeBadRequest,
		Code:    code,
		Message: message,
		Source:  nil,
	}
}

var (
	// ErrUserExisted is returned when a user provides an email existed in database.
	ErrUserExisted error = NewBadRequestError("USER_EXISTED", "user is already existed")

	// ErrUserInvalidEmailAddress represents an error when the provided email format is invalid.
	ErrUserInvalidEmailAddress error = NewBadRequestError("USER_INVALID_EMAIL", "provided user email address is invalid")

	// ErrUserRoleIsEmpty represents an error when the provided user's role format is invalid.
	ErrUserRoleIsEmpty error = NewBadRequestError("USER_INVALID_ROLE", "please provide valid user's role")

	// ErrUserFirstNameEmpty represents an error when the provided name format is empty.
	ErrUserFirstNameEmpty error = NewBadRequestError("USER_INVALID_NAME", "please provide valid user's name")

	// ErrUserPasswordEmpty represents an error when the provided password is empty.
	ErrUserPasswordEmpty error = NewBadRequestError("USER_INVALID_PASSWORD", "please provide valid user's password")

	// ErrInvalidSignInPayload represents an error when the provided email and password combination is not valid.
	ErrInvalidSignInPayload error = NewBadRequestError("AUTH_INVALID_CREDENTIAL", "please provide valid email and password combination")
)
