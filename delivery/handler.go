package delivery

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
)

// RootHandler returns a basic API health check response.
func RootHandler(dbCoonn *sql.DB) echo.HandlerFunc {
	databaseStatus := "ok"
	if err := dbCoonn.Ping(); err != nil {
		databaseStatus = err.Error()
	}

	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, Response{
			StatusCode: http.StatusOK,
			Message:    "API is alive...",
			Data: map[string]any{
				"database": databaseStatus,
			},
		})
	}
}
