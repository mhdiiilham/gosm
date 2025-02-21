package delivery

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	"github.com/mhdiiilham/gosm/entity"
	"github.com/mhdiiilham/gosm/logger"
)

// EventService defines the service interface for event-related operations.
type EventService interface {
	CreateEvent(ctx context.Context, userID string, eventRequest entity.Event) (createdEvent *entity.Event, err error)
	GetEvent(ctx context.Context, userID, UUID string) (event *entity.Event, err error)
	GetEvents(ctx context.Context, userID string, request entity.PaginationRequest) (response entity.PaginationResponse, err error)
	AddGuests(ctx context.Context, userID string, guestList []entity.Guest) (numberOfSuccess int, err error)
	DeleteGuests(ctx context.Context, userID string, guestUUIDs []string) (err error)
	UpdateGuestVIPStatus(ctx context.Context, guestUUID string, vipStatus bool) (err error)
	UpdateEvent(ctx context.Context, event entity.Event) (err error)
	SendGuestInvitation(ctx context.Context, userID, eventUUID, guestUUID string) (status string, err error)
	GetGuestByShortID(ctx context.Context, guestShortID string) (guest *entity.Guest, err error)
	UpdateGuestAttendingStatus(ctx context.Context, guestShortID string, isAttending bool, message string) (err error)
	DeleteEvent(ctx context.Context, eventUUID string) (success bool, err error)
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

	eventDetailGrouped := e.Group("/:uuid")
	eventDetailGrouped.GET("", middleware.AuthMiddleware(AllowedAuthenticatedOnly, h.handleGetEvent))
	eventDetailGrouped.PATCH("", middleware.AuthMiddleware(AllowedAuthenticatedOnly, h.handleUpdateEvent))
	eventDetailGrouped.DELETE("", middleware.AuthMiddleware(AllowedSuperAdminOnly, h.handleDeleteEvent))

	eventDetailedGuestGrouped := eventDetailGrouped.Group("/guests")
	eventDetailedGuestGrouped.POST("", middleware.AuthMiddleware(AllowedAuthenticatedOnly, h.handleAddGuestToEvent))
	eventDetailedGuestGrouped.DELETE("", middleware.AuthMiddleware(AllowedAuthenticatedOnly, h.handleDeleteGuests))
	eventDetailedGuestGrouped.PATCH("/:guest_uuid", middleware.AuthMiddleware(AllowedAuthenticatedOnly, h.handleUpdateGuestVIPStatus))
	eventDetailedGuestGrouped.POST("/:guest_uuid/invite", middleware.AuthMiddleware(AllowedAuthenticatedOnly, h.handleSentInvitation))
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

	userID := c.Get("user_id").(string)

	if err := c.Bind(&request); err != nil {
		logger.Warn(ctx, ops, "failed to parse request body")
		return c.JSON(http.StatusInternalServerError, Response{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal Server Error",
			Data:       nil,
			Error:      err,
		})
	}

	createdEvent, serviceErr := h.eventService.CreateEvent(ctx, userID, entity.Event{
		Name:                 request.Name,
		Host:                 pointer.ToString(request.Host),
		EventType:            entity.ParseEventType(request.EventType),
		Location:             request.Location,
		StartDate:            request.StartDate,
		EndDate:              request.EndDate,
		DigitalInvitationURL: request.DigitalInvitationURL,
		MessageTemplate:      pointer.To(request.MessageTemplate),
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
		Message:    fmt.Sprintf("event %s created", createdEvent.Name),
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
//	@Param			uuid			path		string	true	"Event UUID"
//	@Success		200				{object}	Response{data=entity.Event}
//	@Failure		400				{object}	Response	"Bad Request"
//	@Failure		404				{object}	Response	"Event Not Found"
//	@Failure		500				{object}	Response	"Internal Server Error"
//	@Router			/events/{uuid} [get]
func (h *EventHandler) handleGetEvent(c echo.Context) error {
	ctx := c.Request().Context()
	eventUUID := c.Param("uuid")
	userID := c.Get("user_id").(string)

	event, serviceErr := h.eventService.GetEvent(ctx, userID, eventUUID)
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
			Message:    fmt.Sprintf("event %s not found", eventUUID),
			Data:       nil,
			Error:      nil,
		})
	}

	return c.JSON(http.StatusOK, Response{
		StatusCode: http.StatusOK,
		Message:    fmt.Sprintf("success get event: %s", event.Name),
		Data:       event,
		Error:      nil,
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
	userID := c.Get("user_id").(string)
	page := 1
	perPage := 10

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

	if c.QueryParam("per_page") != "" {
		perPage, err = strconv.Atoi(c.QueryParam("per_page"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, Response{
				StatusCode: http.StatusBadRequest,
				Message:    "invalid query parameter 'per_page' value.",
				Data:       nil,
				Error:      nil,
			})
		}
	}

	eventPaginatedResponse, err := h.eventService.GetEvents(ctx, userID, entity.PaginationRequest{Page: page, PerPage: perPage})
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
//	@Router			/events/{uuid}/guests [post]
func (h *EventHandler) handleAddGuestToEvent(c echo.Context) error {
	var request AddGuestRequest
	ctx := c.Request().Context()
	eventUUID := c.Param("uuid")

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	var guestList []entity.Guest
	for _, guest := range request.Guests {
		toBeAddedGuest := entity.Guest{
			Name:        guest.Name,
			PhoneNumber: guest.PhoneNumber,
			IsVIP:       guest.IsVIP,
		}

		if err := toBeAddedGuest.AssignShortID(); err != nil {
			// TODO: handle error later, for now just log it.
			logger.Errorf(ctx, "Event.Handler", "failed to assign guest's short id; err: %v", err)
		}
		guestList = append(guestList, toBeAddedGuest)
	}

	numberOfSuccess, err := h.eventService.AddGuests(ctx, eventUUID, guestList)
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
	userID := c.Get("user_id").(string)

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	var targetDeleteUUIDs []string
	for _, guest := range request.Guests {
		targetDeleteUUIDs = append(targetDeleteUUIDs, guest.GuestUUID)
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
//	@Param			guest_uuid		path		string		true	"Guest UUID"
//	@Param			is_vip			query		boolean		true	"VIP status (true/false)"
//	@Success		200				{object}	Response	"Guest VIP status updated successfully"
//	@Failure		400				{object}	Response	"Bad Request"
//	@Failure		500				{object}	Response	"Internal Server Error"
//	@Router			/guests/{guest_uuid} [patch]
func (h *EventHandler) handleUpdateGuestVIPStatus(c echo.Context) error {
	ctx := c.Request().Context()
	isVIPQueryParam := c.QueryParam("is_vip")
	guestUUID := c.Param("guest_uuid")

	isVIP, _ := strconv.ParseBool(isVIPQueryParam)

	if err := h.eventService.UpdateGuestVIPStatus(ctx, guestUUID, isVIP); err != nil {
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
//	@Param			uuid			path		string				true	"Event UUID"
//	@Param			body			body		CreateEventRequest	true	"Event update payload"
//	@Success		200				{object}	Response			"Event updated successfully"
//	@Failure		400				{object}	Response			"Bad Request"
//	@Failure		500				{object}	Response			"Internal Server Error"
//	@Router			/events/{uuid} [patch]
func (h *EventHandler) handleUpdateEvent(c echo.Context) error {
	eventUUID := c.Param("uuid")
	ctx := c.Request().Context()
	var request CreateEventRequest

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	if err := h.eventService.UpdateEvent(ctx, entity.Event{
		UUID:                 eventUUID,
		Name:                 request.Name,
		Host:                 pointer.ToString(request.Host),
		EventType:            entity.ParseEventType(request.EventType),
		Location:             request.Location,
		StartDate:            request.StartDate,
		EndDate:              request.EndDate,
		DigitalInvitationURL: request.DigitalInvitationURL,
		MessageTemplate:      pointer.To(request.MessageTemplate),
	}); err != nil {
		return c.JSON(http.StatusInternalServerError, throwInternalServerError(err))
	}

	return c.JSON(http.StatusOK, Response{StatusCode: http.StatusOK, Message: "updated!"})
}

// handleSentInvitation sends an invitation to a guest for a specific event.
//
//	@Summary		Send guest invitation
//	@Description	Allows authenticated users to send invitations to guests for an event.
//	@Tags			invitations
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string		true	"Bearer Token"
//	@Param			uuid			path		string		true	"Event UUID"
//	@Param			guest_uuid		path		string		true	"Guest UUID"
//	@Success		200				{object}	Response	"Invitation sent successfully"
//	@Failure		400				{object}	Response	"Bad Request"
//	@Failure		500				{object}	Response	"Internal Server Error"
//	@Router			/events/{uuid}/guests/{guest_uuid}/invite [post]
func (h *EventHandler) handleSentInvitation(c echo.Context) error {
	guestUUID := c.Param("guest_uuid")
	eventUUID := c.Param("uuid")
	userID := c.Get("user_id").(string)

	ctx := c.Request().Context()
	status, err := h.eventService.SendGuestInvitation(ctx, userID, eventUUID, guestUUID)
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
		Message:    fmt.Sprintf("Invitation sent, status: %s", status),
		Data:       nil,
		Error:      nil,
	})
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
	eventUUID := c.Param("uuid")
	ctx := c.Request().Context()

	success, err := h.eventService.DeleteEvent(ctx, eventUUID)
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
		Message:    fmt.Sprintf("Success delete event %s: %v", eventUUID, success),
		Data:       nil,
		Error:      nil,
	})
}
