package delivery

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// GetGuestByItShortID retrieves guest information using a short ID.
//
//	@Summary		Get guest by short ID
//	@Description	Fetches guest information without requiring authentication.
//	@Tags			guests
//	@Accept			json
//	@Produce		json
//	@Param			short_id	query		string						true	"Guest Short ID"
//	@Success		200			{object}	Response{data=entity.Guest}	"Successfully retrieved guest"
//	@Failure		400			{object}	Response					"Bad Request"
//	@Failure		404			{object}	Response					"Guest not found"
//	@Failure		500			{object}	Response					"Internal Server Error"
//	@Router			/api/v1/public/guests [get]
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

// UpdateGuestAttendingFromInvitation updates a guest's attending status.
//
//	@Summary		Update guest attending status
//	@Description	Allows guests to update their attending status using a short ID.
//	@Tags			guests
//	@Accept			json
//	@Produce		json
//	@Param			request	body		UpdateGuestAttendingAndMessage	true	"Guest attending status update request"
//	@Success		200		{object}	Response						"Successfully updated guest status"
//	@Failure		400		{object}	Response						"Bad Request"
//	@Failure		500		{object}	Response						"Internal Server Error"
//	@Router			/api/v1/public/guests [post]
func UpdateGuestAttendingFromInvitation(srv EventService) echo.HandlerFunc {
	return func(c echo.Context) error {
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
