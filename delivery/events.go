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
	e.GET("", middleware.AuthMiddleware([]entity.UserRole{entity.UserRoleSuperAdmin, entity.UserRoleHost, entity.UserRoleOrganizer, entity.UserRoleGuest}, h.handleGetEvents))
	e.POST("", middleware.AuthMiddleware([]entity.UserRole{"super_admin"}, h.handleCreateEvent))

	eventDetailGrouped := e.Group("/:uuid")
	eventDetailGrouped.GET("", middleware.AuthMiddleware([]entity.UserRole{entity.UserRoleSuperAdmin, entity.UserRoleHost, entity.UserRoleOrganizer, entity.UserRoleGuest}, h.handleGetEvent))
	eventDetailGrouped.PATCH("", middleware.AuthMiddleware([]entity.UserRole{entity.UserRoleSuperAdmin, entity.UserRoleHost, entity.UserRoleOrganizer, entity.UserRoleGuest}, h.handleUpdateEvent))

	eventDetailedGuestGrouped := eventDetailGrouped.Group("/guests")
	eventDetailedGuestGrouped.POST("", middleware.AuthMiddleware([]entity.UserRole{entity.UserRoleSuperAdmin, entity.UserRoleHost, entity.UserRoleOrganizer}, h.handleAddGuestToEvent))
	eventDetailedGuestGrouped.DELETE("", middleware.AuthMiddleware([]entity.UserRole{entity.UserRoleSuperAdmin, entity.UserRoleHost, entity.UserRoleOrganizer}, h.handleDeleteGuests))
	eventDetailedGuestGrouped.PATCH("/:guest_uuid", middleware.AuthMiddleware([]entity.UserRole{entity.UserRoleSuperAdmin, entity.UserRoleHost, entity.UserRoleOrganizer}, h.handleUpdateGuestVIPStatus))
	eventDetailedGuestGrouped.POST("/:guest_uuid/invite", middleware.AuthMiddleware([]entity.UserRole{entity.UserRoleSuperAdmin, entity.UserRoleHost, entity.UserRoleOrganizer}, h.handleSentInvitation))
}

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
