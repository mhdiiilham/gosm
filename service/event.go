package service

import (
	"context"
	"database/sql"

	"github.com/mhdiiilham/gosm/entity"
	"github.com/mhdiiilham/gosm/logger"
)

// EventRepository defines the contract for event-related database operations.
type EventRepository interface {
	CreateEvent(ctx context.Context, event entity.Event) (createdEvent *entity.Event, err error)
	GetEvent(ctx context.Context, tx *sql.Tx, userID, eventID int) (event *entity.Event, err error)
	GetEvents(ctx context.Context, companyID int, limit, offset int) ([]entity.Event, int, error)
	AddGuests(ctx context.Context, eventID int, guestList []entity.Guest) (numberOfSuccess int, err error)
	GetGuests(ctx context.Context, eventID int) (response []entity.Guest, err error)
	DeleteGuests(ctx context.Context, userID int, guestIDs []int) error
	UpdateGuestVIPStatus(ctx context.Context, guestID int, vipStatus bool) error
	GetGuest(ctx context.Context, barcodeID string) (guest *entity.Guest, err error)
	UpdateEvent(ctx context.Context, event entity.Event) (err error)
	UpdateGuestInvitation(ctx context.Context, guest entity.Guest) (err error)
	UpdateGuestAttendingStatus(ctx context.Context, guestID int, isAttending bool, message string) (err error)
	DeleteEvent(ctx context.Context, eventID int) (bool, error)
	SetGuestIsArrived(ctx context.Context, guestID int, isArrived bool) (err error)
}

// KirimWAClient defines an interface for sending WhatsApp messages.
type KirimWAClient interface {
	SendMessage(ctx context.Context, destination string, message string) (id, status string, err error)
}

// EventService provides business logic for managing events.
type EventService struct {
	eventRepository         EventRepository
	kirimWAClient           KirimWAClient
	eventRepositoryRunTxFun func(ctx context.Context, fn entity.TransactionFunc) error
}

// NewEventService initializes a new EventService with a given EventRepository.
func NewEventService(
	eventRepository EventRepository,
	kirimWAClient KirimWAClient,
	eventRepositoryRunTxFun func(ctx context.Context, fn entity.TransactionFunc) error,
) *EventService {
	return &EventService{
		eventRepository:         eventRepository,
		kirimWAClient:           kirimWAClient,
		eventRepositoryRunTxFun: eventRepositoryRunTxFun,
	}
}

// CreateEvent handles the creation of a new event.
// It generates a unique UUID for the event before persisting it.
func (s *EventService) CreateEvent(ctx context.Context, eventRequest entity.Event) (createdEvent *entity.Event, err error) {
	const ops = "EventService.CreateEvent"

	createdEvent, err = s.eventRepository.CreateEvent(ctx, eventRequest)
	if err != nil {
		logger.Errorf(ctx, ops, "failed to create event: %v", err)
		return nil, entity.UnknownError(err)
	}

	return createdEvent, nil
}

// GetEvent retrieves a specific event for a user based on the event UUID.
// If no event is found, it returns nil without an error.
func (s *EventService) GetEvent(ctx context.Context, userID, eventID int) (*entity.Event, error) {
	const ops = "EventService.GetEvent"

	var targetEvent *entity.Event
	if err := s.eventRepositoryRunTxFun(ctx, func(ctx context.Context, tx *sql.Tx) error {
		event, err := s.eventRepository.GetEvent(ctx, tx, userID, eventID)
		if err != nil {
			logger.Errorf(ctx, ops, "failed to get event: %v", err)
			return err
		}
		targetEvent = event
		return nil
	}); err != nil {
		return nil, err
	}

	return targetEvent, nil
}

// GetEvents retrieves a paginated list of events for a specific company.
func (s *EventService) GetEvents(ctx context.Context, companyID int, request entity.PaginationRequest) (response entity.PaginationResponse, err error) {
	const ops = "EventService.GetEvents"

	offset := (request.Page - 1) * request.PerPage
	events, totalRecords, err := s.eventRepository.GetEvents(ctx, companyID, request.PerPage, offset)
	if err != nil {
		logger.Errorf(ctx, ops, "failed to get events")
		return response, err
	}

	lastPage := (totalRecords + request.PerPage - 1) / request.PerPage

	return entity.PaginationResponse{
		Records:      events,
		Page:         request.Page,
		PerPage:      request.PerPage,
		LastPage:     int(lastPage),
		TotalRecords: totalRecords,
	}, nil
}

// AddGuests insert multple of guest into an event.
func (s *EventService) AddGuests(ctx context.Context, eventID int, guestList []entity.Guest) (numberOfSuccess int, err error) {
	return s.eventRepository.AddGuests(ctx, eventID, guestList)
}

// DeleteGuests deletes list of selected guests.
func (s *EventService) DeleteGuests(ctx context.Context, userID int, guestIDs []int) (err error) {
	return s.eventRepository.DeleteGuests(ctx, userID, guestIDs)
}

// UpdateGuestVIPStatus update the guest's vip status.
func (s *EventService) UpdateGuestVIPStatus(ctx context.Context, guestIDs int, vipStatus bool) (err error) {
	return s.eventRepository.UpdateGuestVIPStatus(ctx, guestIDs, vipStatus)
}

// UpdateEvent update given event.
func (s *EventService) UpdateEvent(ctx context.Context, event entity.Event) (err error) {
	return s.eventRepository.UpdateEvent(ctx, event)
}

// SendGuestInvitation sends an invitation message to a guest for a specific event.
func (s *EventService) SendGuestInvitation(ctx context.Context, userID, eventUUID, guestUUID string) (status string, err error) {
	return "", nil
}

// UpdateGuestAttendingStatus update guest's attending status and message.
func (s *EventService) UpdateGuestAttendingStatus(ctx context.Context, guestID int, isAttending bool, message string) (err error) {
	return s.eventRepository.UpdateGuestAttendingStatus(ctx, guestID, isAttending, message)
}

// DeleteEvent soft delete an event based on given event uuid.
func (s *EventService) DeleteEvent(ctx context.Context, eventID int) (success bool, err error) {
	return s.eventRepository.DeleteEvent(ctx, eventID)
}

// SetGuestIsArrived set an guest is_arrived status.
func (s *EventService) SetGuestIsArrived(ctx context.Context, guestID int, isArrived bool) (err error) {
	return s.eventRepository.SetGuestIsArrived(ctx, guestID, isArrived)
}

// GetGuests ...
func (s *EventService) GetGuests(ctx context.Context, eventID int) (guests []entity.Guest, err error) {
	return s.eventRepository.GetGuests(ctx, eventID)
}

// GetGuest ...
func (s *EventService) GetGuest(ctx context.Context, barcodeID string) (guest *entity.Guest, err error) {
	return s.eventRepository.GetGuest(ctx, barcodeID)
}
