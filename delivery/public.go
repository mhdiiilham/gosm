package delivery

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/mhdiiilham/gosm/entity"
	"github.com/mhdiiilham/gosm/pkg"
)

// CountryService defines the interface for country-related data operations.
type CountryService interface {
	GetCountries(ctx context.Context) (countries []entity.Country, err error)
}

// GetGuestByItShortID retrieves guest information using a short ID.
//
//	@Summary		Get guest by short ID
//	@Description	Fetches guest information without requiring authentication.
//	@Tags			public
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
		guestID := c.QueryParam("id")

		ctx := c.Request().Context()
		guest, err := srv.GetGuest(ctx, guestID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
		}

		return c.JSON(http.StatusOK, Response{
			StatusCode: http.StatusOK,
			Message:    fmt.Sprintf("success fetch guest: %s", guestID),
			Data:       guest,
			Error:      nil,
		})
	}
}

// HandleGetCountries handles the request to retrieve a list of countries.
// TODO: Need to move this to gRPC server.
//
//	@Summary		Get list of countries
//	@Description	Fetches a list of countries with their names, flags, and phone international prefixes.
//	@Tags			public
//	@Produce		json
//
//	@Success		200	{object}	Response{data=GetCountriesResponse}	"Successfully get countries"
//	@Failure		500	{object}	Response							"Internal Server Error"
//
//	@Router			/api/v1/public/countries [get]
func HandleGetCountries(srv CountryService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		countries, err := srv.GetCountries(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
		}

		return c.JSON(
			http.StatusOK,
			Response{
				StatusCode: http.StatusOK,
				Message:    "success get countries",
				Data: GetCountriesResponse{
					Countries: countries,
				},
				Error: nil,
			},
		)

	}
}

func AddGuestToEvent(srv EventService) echo.HandlerFunc {
	return func(c echo.Context) error {
		eventID, _ := strconv.Atoi(c.Param("eventId"))

		var request PublicAddGuestRequest
		if err := c.Bind(&request); err != nil {
			return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
		}

		if request.ID != "" {
			if err := srv.UpdateGuest(c.Request().Context(), request.ID, request.Name, pkg.FormatPhoneToWaMe(request.Phone), request.Message, request.IsAttending); err != nil {
				return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
			}

			return c.JSON(http.StatusOK, Response{
				StatusCode: http.StatusOK,
				Message:    "guest added",
				Data:       request.ID,
				Error:      nil,
			})
		}

		barcodeID, _ := pkg.GeneratePumBookID(strconv.Itoa(eventID))
		_, err := srv.AddGuests(c.Request().Context(), eventID, []entity.Guest{
			{

				EventID:     eventID,
				BarcodeID:   barcodeID,
				Name:        request.Name,
				Phone:       pkg.FormatPhoneToWaMe(request.Phone),
				IsAttending: request.IsAttending,
				Message:     request.Message,
			},
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
		}

		return c.JSON(http.StatusOK, Response{
			StatusCode: http.StatusOK,
			Message:    "guest added",
			Data:       barcodeID,
			Error:      nil,
		})
	}
}

func HandleGetGuestMessages(srv EventService) echo.HandlerFunc {
	return func(c echo.Context) error {
		eventID := c.Param("eventId")

		messages, _ := srv.GetGuestMessages(c.Request().Context(), eventID)

		return c.JSON(http.StatusOK, Response{
			StatusCode: http.StatusOK,
			Message:    "",
			Data:       messages,
			Error:      nil,
		})
	}
}
