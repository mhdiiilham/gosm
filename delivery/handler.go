package delivery

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// RootHandler returns a basic API health check response.
func RootHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, Response{
			StatusCode: http.StatusOK,
			Message:    "API is alive...",
			Data:       map[string]string{"foo": "bar"}, // Proper JSON object instead of a raw string
		})
	}
}
