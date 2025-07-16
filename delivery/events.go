package delivery

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/labstack/echo/v4"
	"github.com/mhdiiilham/gosm/entity"
	"github.com/mhdiiilham/gosm/logger"
	"github.com/mhdiiilham/gosm/pkg"
)

// EventService defines the service interface for event-related operations.
type EventService interface {
	CreateEvent(ctx context.Context, eventRequest entity.Event) (createdEvent *entity.Event, err error)
	GetEvent(ctx context.Context, userID, EventID int) (event *entity.Event, err error)
	GetEvents(ctx context.Context, userID int, request entity.PaginationRequest) (response entity.PaginationResponse, err error)
	AddGuests(ctx context.Context, eventID int, guestList []entity.Guest) (numberOfSuccess int, err error)
	DeleteGuests(ctx context.Context, userID int, guestIDs []int) (err error)
	UpdateGuestVIPStatus(ctx context.Context, guestID int, vipStatus bool) (err error)
	UpdateEvent(ctx context.Context, event entity.Event) (err error)
	DeleteEvent(ctx context.Context, eventID int) (success bool, err error)
	SetGuestIsArrived(ctx context.Context, guestID int, isArrived bool) (err error)
	GetGuests(ctx context.Context, eventID int) (guests []entity.Guest, err error)
	GetGuest(ctx context.Context, barcodeID string) (guest *entity.Guest, err error)
	UpdateGuest(ctx context.Context, guestID, name, phone, message string, isAttending bool) error
	GetGuestMessages(ctx context.Context, eventID string) ([]entity.GuestMessages, error)
}

// EventHandler handles HTTP requests related to event operations.
type EventHandler struct {
	eventService EventService
}

// NewEventHandler creates a new instance of EventHandler.
func NewEventHandler(service EventService) *EventHandler {
	return &EventHandler{eventService: service}
}

// RegisterEventRoutes registers the event-related routes within the Echo router group.
func (h *EventHandler) RegisterEventRoutes(e *echo.Group, middleware *Middleware) {
	e.GET("", middleware.AuthMiddleware(AllowedAuthenticatedOnly, h.handleGetEvents))
	e.POST("", middleware.AuthMiddleware(AllowedAuthenticatedOnly, h.handleCreateEvent))

	eventDetailGrouped := e.Group("/:id")
	eventDetailGrouped.GET("", middleware.AuthMiddleware(AllowedAuthenticatedOnly, h.handleGetEvent))
	eventDetailGrouped.PATCH("", middleware.AuthMiddleware(AllowedAuthenticatedOnly, h.handleUpdateEvent))
	eventDetailGrouped.DELETE("", middleware.AuthMiddleware(AllowedSuperAdminOnly, h.handleDeleteEvent))

	eventDetailedGuestGrouped := eventDetailGrouped.Group("/guests")
	eventDetailedGuestGrouped.GET("", middleware.AuthMiddleware(AllowedAuthenticatedOnly, h.handleGetGuests))
	eventDetailedGuestGrouped.POST("", middleware.AuthMiddleware(AllowedAuthenticatedOnly, h.handleAddGuestToEvent))
	eventDetailedGuestGrouped.POST("/csv", middleware.AuthMiddleware(AllowedAuthenticatedOnly, h.handleAddGuestCSV))
	eventDetailedGuestGrouped.POST("/arrived", h.handleUpdateGuestArrived)
	eventDetailedGuestGrouped.DELETE("", middleware.AuthMiddleware(AllowedAuthenticatedOnly, h.handleDeleteGuests))
}

// @Summary		Create an event
// @Description	Creates a new event for the authenticated user.
// @Tags			events
// @Accept			json
// @Produce		json
// @Security		BearerAuth
// @Param			Authorization	header		string				true	"Bearer Token"
// @Param			request			body		CreateEventRequest	true	"Event creation payload"
// @Success		201				{object}	Response{data=entity.Event}
// @Failure		400				{object}	Response	"Bad Request"
// @Failure		500				{object}	Response	"Internal Server Error"
// @Router			/events [post]
func (h *EventHandler) handleCreateEvent(c echo.Context) error {
	ctx := c.Request().Context()
	const ops = "EventHandler.handleCreateEvent"
	var request CreateEventRequest

	userID := c.Get("user_id").(int)
	companyID := c.Get("company_id").(int)

	if err := c.Bind(&request); err != nil {
		logger.Warn(ctx, ops, "failed to parse request body")
		return c.JSON(http.StatusInternalServerError, Response{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal Server Error",
			Data:       nil,
			Error:      err,
		})
	}

	createdEvent, serviceErr := h.eventService.CreateEvent(ctx, entity.Event{
		Title:       request.Title,
		Type:        entity.ParseEventType(request.Type),
		Description: request.Description,
		Location:    request.Location,
		StartDate:   request.StartDate,
		EndDate:     request.EndDate,
		CreatedBy: entity.IDName{
			ID: userID,
		},
		Company: entity.IDName{
			ID: companyID,
		},
		GuestCount: request.GuestCount,
	})

	if serviceErr != nil {
		switch err := serviceErr.(type) {
		case entity.GosmError:
			if err.Type == entity.GosmErrorTypeBadRequest {
				return c.JSON(http.StatusBadRequest, Response{
					StatusCode: http.StatusBadRequest,
					Message:    err.Message,
					Data:       nil,
					Error:      err.Source,
				})
			}
		}

		return c.JSON(http.StatusInternalServerError, throwInternalServerError(serviceErr))
	}

	return c.JSON(http.StatusCreated, Response{
		StatusCode: http.StatusCreated,
		Message:    fmt.Sprintf("event %s created", createdEvent.Title),
		Data:       createdEvent,
		Error:      nil,
	})
}

// handleGetEvent retrieves an event by UUID.
//
//	@Summary		Get an event
//	@Description	Fetches event details for the authenticated user.
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string	true	"Bearer Token"
//	@Param			id				path		string	true	"Event UUID"
//	@Success		200				{object}	Response{data=entity.Event}
//	@Failure		400				{object}	Response	"Bad Request"
//	@Failure		404				{object}	Response	"Event Not Found"
//	@Failure		500				{object}	Response	"Internal Server Error"
//	@Router			/events/{id} [get]
func (h *EventHandler) handleGetEvent(c echo.Context) error {
	ctx := c.Request().Context()
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	userID := c.Get("user_id").(int)

	event, serviceErr := h.eventService.GetEvent(ctx, userID, eventID)
	if serviceErr != nil {
		switch err := serviceErr.(type) {
		case entity.GosmError:
			if err.Type == entity.GosmErrorTypeBadRequest {
				return c.JSON(http.StatusBadRequest, Response{
					StatusCode: http.StatusBadRequest,
					Message:    err.Message,
					Data:       nil,
					Error:      err.Source,
				})
			}
		}

		return c.JSON(http.StatusInternalServerError, throwInternalServerError(serviceErr))
	}

	if event == nil {
		return c.JSON(http.StatusNotFound, Response{
			StatusCode: http.StatusNotFound,
			Message:    fmt.Sprintf("event %d not found", eventID),
			Data:       nil,
			Error:      nil,
		})
	}

	status := "Upcoming"
	if time.Now().After(event.EndDate) {
		status = "Past"
	}

	return c.JSON(http.StatusOK, Response{
		StatusCode: http.StatusOK,
		Message:    fmt.Sprintf("success get event: %s", event.Title),
		Data: EventResponse{
			ID:             event.ID,
			Name:           event.Title,
			Type:           string(event.Type),
			StartDate:      event.StartDate.Format(time.RFC3339),
			EndDate:        event.EndDate.Format(time.RFC3339),
			Location:       event.Location,
			Description:    event.Description,
			GuestCount:     event.GuestCount,
			CheckedInCount: 0,
			Status:         status,
		},
		Error: nil,
	})
}

// handleGetEvents retrieves a paginated list of events for the authenticated user.
//
//	@Summary		Get list of events
//	@Description	Fetches a paginated list of events for the authenticated user.
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string	true	"Bearer Token"
//	@Param			name			query		string	false	"Event name"
//	@Param			host			query		string	false	"Event host"
//	@Param			page			query		int		false	"Page number (default: 1)"
//	@Param			per_page		query		int		false	"Items per page (default: 10)"
//	@Success		200				{object}	Response{data=entity.PaginationResponse{data=[]entity.Event}}
//	@Failure		400				{object}	Response	"Bad Request"
//	@Failure		500				{object}	Response	"Internal Server Error"
//	@Router			/events [get]
func (h *EventHandler) handleGetEvents(c echo.Context) error {
	ctx := c.Request().Context()
	companyID := c.Get("company_id").(int)
	page := 1

	var err error
	if c.QueryParam("page") != "" {
		page, err = strconv.Atoi(c.QueryParam("page"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, Response{
				StatusCode: http.StatusBadRequest,
				Message:    "invalid query parameter 'page' value.",
				Data:       nil,
				Error:      nil,
			})
		}
	}

	eventPaginatedResponse, err := h.eventService.GetEvents(ctx, companyID, entity.PaginationRequest{Page: page, PerPage: math.MaxInt})
	if err != nil {
		switch parsedErr := err.(type) {
		case entity.GosmError:
			if parsedErr.Type == entity.GosmErrorTypeBadRequest {
				return c.JSON(http.StatusBadRequest, Response{
					StatusCode: http.StatusBadRequest,
					Message:    parsedErr.Message,
					Data:       nil,
					Error:      parsedErr.Source,
				})
			}
		}

		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	var events []EventResponse
	for _, event := range eventPaginatedResponse.Records.([]entity.Event) {
		status := "Upcoming"
		if time.Now().After(event.EndDate) {
			status = "Past"
		}

		events = append(events, EventResponse{
			ID:          event.ID,
			Name:        event.Title,
			Type:        string(event.Type),
			Description: event.Description,
			Location:    event.Location,
			StartDate:   event.StartDate.Format(time.RFC3339),
			EndDate:     event.EndDate.Format(time.RFC3339),
			GuestCount:  event.GuestCount,
			Status:      status,
		})

	}

	eventPaginatedResponse.Records = events

	return c.JSON(http.StatusOK, Response{
		StatusCode: http.StatusOK,
		Message:    "success",
		Data:       eventPaginatedResponse,
		Error:      nil,
	})
}

// handleAddGuestToEvent adds guests to an event.
//
//	@Summary		Add guests to an event
//	@Description	Allows authenticated users to add multiple guests to a specific event.
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string			true	"Bearer Token"
//	@Param			uuid			path		string			true	"Event UUID"
//	@Param			request			body		AddGuestRequest	true	"List of guests to be added"
//	@Success		200				{object}	Response		"Success message with number of guests added"
//	@Failure		400				{object}	Response		"Bad Request"
//	@Failure		500				{object}	Response		"Internal Server Error"
//	@Router			/events/{id}/guests [post]
func (h *EventHandler) handleAddGuestToEvent(c echo.Context) error {
	var request AddGuestRequest
	ctx := c.Request().Context()
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	var guestList []entity.Guest
	for _, guest := range request.Guests {
		toBeAddedGuest := entity.Guest{
			Name:  guest.Name,
			Email: guest.Email,
			Phone: pkg.FormatPhoneToWaMe(guest.PhoneNumber),
			IsVIP: guest.IsVIP,
		}

		guestList = append(guestList, toBeAddedGuest)
	}

	numberOfSuccess, err := h.eventService.AddGuests(ctx, eventID, guestList)
	if err != nil {
		switch parsedErr := err.(type) {
		case entity.GosmError:
			if parsedErr.Type == entity.GosmErrorTypeBadRequest {
				return c.JSON(http.StatusBadRequest, Response{
					StatusCode: http.StatusBadRequest,
					Message:    parsedErr.Message,
					Data:       nil,
					Error:      parsedErr.Source,
				})
			}
		}

		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	return c.JSON(http.StatusOK, Response{
		StatusCode: http.StatusOK,
		Message:    fmt.Sprintf("success added %d/%d guests to the event", numberOfSuccess, len(request.Guests)),
		Data:       nil,
		Error:      nil,
	})
}

// handleDeleteGuests removes guests from an event.
//
//	@Summary		Delete guests from an event
//	@Description	Allows authenticated users to remove multiple guests from a specific event.
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string			true	"Bearer Token"
//	@Param			request			body		AddGuestRequest	true	"List of guests to be deleted (UUIDs required)"
//	@Success		200				{object}	Response		"Guests successfully removed"
//	@Failure		400				{object}	Response		"Bad Request"
//	@Failure		500				{object}	Response		"Internal Server Error"
//	@Router			/events/guests [delete]
func (h *EventHandler) handleDeleteGuests(c echo.Context) error {
	var request AddGuestRequest
	ctx := c.Request().Context()
	userID := c.Get("user_id").(int)

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	var targetDeleteUUIDs []int
	for _, guest := range request.Guests {
		targetDeleteUUIDs = append(targetDeleteUUIDs, guest.ID)
	}

	if err := h.eventService.DeleteGuests(ctx, userID, targetDeleteUUIDs); err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	return c.JSON(http.StatusOK, Response{StatusCode: http.StatusOK, Message: "ok"})
}

// handleUpdateGuestVIPStatus updates the VIP status of a guest.
//
//	@Summary		Update guest VIP status
//	@Description	Allows authenticated users to change a guest's VIP status.
//	@Tags			guests
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string		true	"Bearer Token"
//	@Param			guest_id		path		string		true	"Guest UUID"
//	@Param			is_vip			query		boolean		true	"VIP status (true/false)"
//	@Success		200				{object}	Response	"Guest VIP status updated successfully"
//	@Failure		400				{object}	Response	"Bad Request"
//	@Failure		500				{object}	Response	"Internal Server Error"
//	@Router			/guests/{guest_id} [patch]
func (h *EventHandler) handleUpdateGuestVIPStatus(c echo.Context) error {
	ctx := c.Request().Context()
	isVIPQueryParam := c.QueryParam("is_vip")
	guestID, err := strconv.Atoi(c.Param("guest_id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	isVIP, _ := strconv.ParseBool(isVIPQueryParam)

	if err := h.eventService.UpdateGuestVIPStatus(ctx, guestID, isVIP); err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	return c.JSON(http.StatusOK, Response{StatusCode: http.StatusOK, Message: "ok"})
}

// handleUpdateEvent updates an existing event.
//
//	@Summary		Update an event
//	@Description	Allows authenticated users to update event details.
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string				true	"Bearer Token"
//	@Param			id				path		string				true	"Event UUID"
//	@Param			body			body		CreateEventRequest	true	"Event update payload"
//	@Success		200				{object}	Response			"Event updated successfully"
//	@Failure		400				{object}	Response			"Bad Request"
//	@Failure		500				{object}	Response			"Internal Server Error"
//	@Router			/events/{id} [patch]
func (h *EventHandler) handleUpdateEvent(c echo.Context) error {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	ctx := c.Request().Context()
	var request CreateEventRequest

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	if err := h.eventService.UpdateEvent(ctx, entity.Event{
		ID:          eventID,
		Title:       request.Title,
		Type:        entity.ParseEventType(request.Type),
		Description: request.Description,
		Location:    request.Location,
		StartDate:   request.StartDate,
		EndDate:     request.EndDate,
	}); err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	return c.JSON(http.StatusOK, Response{StatusCode: http.StatusOK, Message: "updated!"})
}

// handleDeleteEvent deletes an event.
//
//	@Summary		Delete an event
//	@Description	Allows only super admins to delete an event.
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string		true	"Bearer Token"
//	@Param			uuid			path		string		true	"Event UUID"
//	@Success		200				{object}	Response	"Event deleted successfully"
//	@Failure		400				{object}	Response	"Bad Request"
//	@Failure		403				{object}	Response	"Forbidden - Only super admins allowed"
//	@Failure		500				{object}	Response	"Internal Server Error"
//	@Router			/events/{uuid} [delete]
func (h *EventHandler) handleDeleteEvent(c echo.Context) error {
	eventID, err := strconv.Atoi(c.Param("uuid"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	ctx := c.Request().Context()

	success, err := h.eventService.DeleteEvent(ctx, eventID)
	if err != nil {
		switch parsedErr := err.(type) {
		case entity.GosmError:
			if parsedErr.Type == entity.GosmErrorTypeBadRequest {
				return c.JSON(http.StatusBadRequest, Response{
					StatusCode: http.StatusBadRequest,
					Message:    parsedErr.Message,
					Data:       nil,
					Error:      parsedErr.Source,
				})
			}
		}

		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	return c.JSON(http.StatusOK, Response{
		StatusCode: http.StatusOK,
		Message:    fmt.Sprintf("Success delete event %d: %v", eventID, success),
		Data:       nil,
		Error:      nil,
	})
}

// handleUpdateGuestArrived updates the arrival status of a guest.
//
//	@Summary		Update guest arrival status
//	@Description	Updates the arrival status of a guest using their short ID.
//	@Tags			Guests
//	@Accept			json
//	@Produce		json
//	@Param			short_id	query		string		true	"Guest Short ID"
//	@Param			is_arrived	query		bool		true	"Arrival status (true/false)"
//	@Success		200			{object}	Response	"Guest arrival status updated successfully"
//	@Failure		400			{object}	Response	"Bad request (invalid guest ID or parameters)"
//	@Failure		500			{object}	Response	"Internal server error"
//	@Router			/events/{uuid}/guests/arrived [post]
func (h *EventHandler) handleUpdateGuestArrived(c echo.Context) error {
	guestID, err := strconv.Atoi(c.QueryParam("short_id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	isArrived, _ := strconv.ParseBool(c.QueryParam("is_arrived"))
	ctx := c.Request().Context()

	if err := h.eventService.SetGuestIsArrived(ctx, guestID, isArrived); err != nil {
		switch parsedErr := err.(type) {
		case entity.GosmError:
			if parsedErr.Type == entity.GosmErrorTypeBadRequest {
				return c.JSON(http.StatusBadRequest, Response{
					StatusCode: http.StatusBadRequest,
					Message:    parsedErr.Message,
					Data:       nil,
					Error:      parsedErr.Source,
				})
			}
		}

		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	return c.JSON(http.StatusOK, Response{
		StatusCode: http.StatusOK,
		Message:    fmt.Sprintf("Guest %d updated to arrived: %v", guestID, isArrived),
		Data:       nil,
		Error:      nil,
	})
}

func (h *EventHandler) handleGetGuests(c echo.Context) error {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	ctx := c.Request().Context()

	guests, err := h.eventService.GetGuests(ctx, eventID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	return c.JSON(http.StatusOK, Response{
		StatusCode: http.StatusOK,
		Message:    "ok",
		Data:       guests,
		Error:      nil,
	})
}

func (h *EventHandler) handleAddGuestCSV(c echo.Context) error {
	ctx := c.Request().Context()
	eventID, err := strconv.Atoi(c.Param("id"))

	f, err := c.FormFile("guest_file")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	fileExt := strings.ToLower(filepath.Ext(f.Filename))
	src, err := f.Open()
	if err != nil {
		logger.Errorf(ctx, "error", "err: %v", err)
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}
	defer src.Close()

	// Read the data asynchronously
	go func() {
		ctx := context.Background()
		var guests []entity.Guest

		switch fileExt {
		case ".csv":
			guestRaw, err := io.ReadAll(src)
			if err != nil {
				logger.Errorf(ctx, "error", "err: %v", err)
				return
			}
			for _, row := range strings.Split(string(guestRaw), "\r\n")[1:] {
				if row == "" {
					continue
				}
				cols := strings.Split(row, ",")
				vipStatus, _ := strconv.ParseBool(cols[3])
				guests = append(guests, entity.Guest{
					EventID: eventID,
					Name:    cols[0],
					Phone:   pkg.FormatPhoneToWaMe(cols[1]),
					Email:   cols[2],
					IsVIP:   vipStatus,
				})
			}
		case ".xlsx":
			tmpFile, err := os.CreateTemp("", "upload-*.xlsx")
			if err != nil {
				logger.Errorf(ctx, "error", "err: %v", err)
				return
			}
			defer os.Remove(tmpFile.Name())
			io.Copy(tmpFile, src)
			tmpFile.Close()

			xlFile, err := excelize.OpenFile(tmpFile.Name())
			if err != nil {
				logger.Errorf(ctx, "error", "err: %v", err)
				return
			}
			rows, err := xlFile.GetRows("Sheet1")
			if err != nil {
				logger.Errorf(ctx, "error", "err: %v", err)
				return
			}
			for _, row := range rows[1:] {
				vipStatus := false
				if len(row) == 4 {
					vipStatus, _ = strconv.ParseBool(row[3])
				}

				guests = append(guests, entity.Guest{
					EventID: eventID,
					Name:    row[0],
					Phone:   pkg.FormatPhoneToWaMe(row[1]),
					Email:   row[2],
					IsVIP:   vipStatus,
				})
			}
		default:
			logger.Errorf(ctx, "unsupported file type: %s", fileExt)
			return
		}

		logger.Infof(ctx, "EventHandler.handleAddGuestCSV", "processing guest list")
		h.eventService.AddGuests(ctx, eventID, guests)
		logger.Infof(ctx, "EventHandler.handleAddGuestCSV", "done processing guest list")
	}()

	return c.JSON(http.StatusOK, Response{
		StatusCode: http.StatusOK,
		Message:    "Importing guest, This might took a while.",
		Error:      nil,
	})
}
