package delivery

import "net/http"

// Response represents a standard API response structure.
// It includes the status code, message, optional data, and an error message (if any).
type Response struct {
	StatusCode int    `json:"code"`
	Message    string `json:"message"`
	Data       any    `json:"data"`
	Error      error  `json:"error"`
}

// throwInternalServerError creates a Response representing an internal server error.
// This function is used to standardize error responses when unexpected server issues occur.
func throwInternalServerError(err error) Response {
	return Response{
		StatusCode: http.StatusInternalServerError,
		Message:    "INTERNAL_SERVER_ERROR",
		Data:       nil,
		Error:      err,
	}
}
