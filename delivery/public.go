package delivery

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// GetGuestByItShortID public handle to retrieve guest information without any authentication.
func GetGuestByItShortID(srv EventService) echo.HandlerFunc {
	return func(c echo.Context) error {
		guestShortID := c.QueryParam("short_id")

		ctx := c.Request().Context()
		guest, err := srv.GetGuestByShortID(ctx, guestShortID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
		}

		return c.JSON(http.StatusOK, Response{
			StatusCode: http.StatusOK,
			Message:    fmt.Sprintf("success fetch guest: %s", guestShortID),
			Data:       guest,
			Error:      nil,
		})
	}
}

// UpdateGuestAttendingFromInvitation public handle updating guest's attending status.
func UpdateGuestAttendingFromInvitation(srv EventService) echo.HandlerFunc {
	return func(c echo.Context) error {
		// guestShortID := c.QueryParam("short_id")
		// isAttending, _ := strconv.ParseBool(c.QueryParam("attending"))
		var request UpdateGuestAttendingAndMessage
		if err := c.Bind(&request); err != nil {
			return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
		}

		if err := srv.UpdateGuestAttendingStatus(context.Background(), request.ShortID, request.IsAttending, request.Message); err != nil {
			return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
		}

		return c.JSON(http.StatusOK, Response{
			StatusCode: http.StatusOK,
			Message:    fmt.Sprintf("success update guest: %s", request.ShortID),
			Data:       nil,
			Error:      nil,
		})
	}
}
