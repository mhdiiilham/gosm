package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/mhdiiilham/gosm/entity"
	"github.com/mhdiiilham/gosm/logger"
	"golang.org/x/sync/errgroup"
)

// EventRepository defines the contract for event-related database operations.
type EventRepository interface {
	CreateEvent(ctx context.Context, userID string, event entity.Event) (createdEvent *entity.Event, err error)
	GetEvent(ctx context.Context, tx *sql.Tx, userID, eventUUID string) (event *entity.Event, err error)
	GetEvents(ctx context.Context, userID string, limit, offset int) ([]entity.Event, int, error)
	AddGuests(ctx context.Context, eventID string, guestList []entity.Guest) (numberOfSuccess int, err error)
	GetGuests(ctx context.Context, tx *sql.Tx, eventUUID string) (guests []entity.Guest, err error)
	DeleteGuests(ctx context.Context, userID string, guestUUIDs []string) error
	UpdateGuestVIPStatus(ctx context.Context, guestUUID string, vipStatus bool) error
	GetGuest(ctx context.Context, guestUUID string) (guest *entity.Guest, err error)
	UpdateEvent(ctx context.Context, event entity.Event) (err error)
	UpdateGuestInvitation(ctx context.Context, guest entity.Guest) (err error)
	GetGuestByShortID(ctx context.Context, guestShortID string) (guest *entity.Guest, err error)
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
func (s *EventService) CreateEvent(ctx context.Context, userID string, eventRequest entity.Event) (createdEvent *entity.Event, err error) {
	const ops = "EventService.CreateEvent"

	generatedUUID := uuid.NewString()
	eventRequest.UUID = generatedUUID

	createdEvent, err = s.eventRepository.CreateEvent(ctx, userID, eventRequest)
	if err != nil {
		logger.Errorf(ctx, ops, "failed to create event: %v", err)
		return nil, entity.UnknownError(err)
	}

	return createdEvent, nil
}

// GetEvent retrieves a specific event for a user based on the event UUID.
// If no event is found, it returns nil without an error.
func (s *EventService) GetEvent(ctx context.Context, userID, eventUUID string) (*entity.Event, error) {
	const ops = "EventService.GetEvent"

	errsGroup, ctx := errgroup.WithContext(ctx)
	eventCh := make(chan *entity.Event, 1)
	guestListCh := make(chan []entity.Guest, 1)

	// fetch event details
	errsGroup.Go(func() error {
		logger.Infof(ctx, ops, "Fetch event %s detail.", eventUUID)
		return s.eventRepositoryRunTxFun(ctx, func(ctx context.Context, tx *sql.Tx) error {
			event, err := s.eventRepository.GetEvent(ctx, tx, userID, eventUUID)
			if err != nil {
				// TODO: handle error properly later.
				logger.Errorf(ctx, ops, "failed to get event: %v", err)
				if err.Error() == "sql: no rows in result set" {
					eventCh <- nil
					return err
				}
			}
			eventCh <- event
			return nil
		})
	})

	// fetch event's guest list.
	errsGroup.Go(func() error {
		logger.Infof(ctx, ops, "Fetch guest list of event: %s.", eventUUID)
		return s.eventRepositoryRunTxFun(ctx, func(ctx context.Context, tx *sql.Tx) error {
			guests, err := s.eventRepository.GetGuests(ctx, tx, eventUUID)
			if err != nil {
				// TODO: handle error properly later.
				logger.Errorf(ctx, ops, "failed to get guest: %v", err)
				guestListCh <- []entity.Guest{}
				return nil
			}
			guestListCh <- guests
			return nil
		})
	})

	if err := errsGroup.Wait(); err != nil {
		return nil, err
	}

	targetEvent := <-eventCh
	targetEvent.GuestList = []entity.Guest{}
	listOfGuest := <-guestListCh
	targetEvent.GuestList = append(targetEvent.GuestList, listOfGuest...)

	return targetEvent, nil
}

// GetEvents retrieves a paginated list of events for a specific user.
func (s *EventService) GetEvents(ctx context.Context, userID string, request entity.PaginationRequest) (response entity.PaginationResponse, err error) {
	const ops = "EventService.GetEvents"

	offset := (request.Page - 1) * request.PerPage
	events, totalRecords, err := s.eventRepository.GetEvents(ctx, userID, request.PerPage, offset)
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
func (s *EventService) AddGuests(ctx context.Context, eventID string, guestList []entity.Guest) (numberOfSuccess int, err error) {
	return s.eventRepository.AddGuests(ctx, eventID, guestList)
}

// // GetGuests retrieve list of guest of an event.
// func (s *EventService) GetGuests(ctx context.Context, eventID string) (guests []entity.Guest, err error) {
// 	return s.eventRepository.GetGuests(ctx, eventID)
// }

// DeleteGuests deletes list of selected guests.
func (s *EventService) DeleteGuests(ctx context.Context, userID string, guestUUIDs []string) (err error) {
	return s.eventRepository.DeleteGuests(ctx, userID, guestUUIDs)
}

// UpdateGuestVIPStatus update the guest's vip status.
func (s *EventService) UpdateGuestVIPStatus(ctx context.Context, guestUUID string, vipStatus bool) (err error) {
	return s.eventRepository.UpdateGuestVIPStatus(ctx, guestUUID, vipStatus)
}

// UpdateEvent update given event.
func (s *EventService) UpdateEvent(ctx context.Context, event entity.Event) (err error) {
	return s.eventRepository.UpdateEvent(ctx, event)
}

// SendGuestInvitation sends an invitation message to a guest for a specific event.
func (s *EventService) SendGuestInvitation(ctx context.Context, userID, eventUUID, guestUUID string) (status string, err error) {

	// var invitationMessageStatus string
	var eventStartDate time.Time
	var targetEvent *entity.Event
	var targetGuest *entity.Guest
	err = s.eventRepositoryRunTxFun(ctx, func(ctx context.Context, tx *sql.Tx) error {

		// 1. get event digital invitation url
		queriedEvent, err := s.eventRepository.GetEvent(ctx, tx, userID, eventUUID)
		if err != nil {
			return nil
		}
		eventStartDate, _ = time.Parse(time.RFC3339, queriedEvent.StartDate)
		targetEvent = queriedEvent

		// 2. get guest phone number
		queriedGuest, err := s.eventRepository.GetGuest(ctx, guestUUID)
		if err != nil {
			return nil
		}

		// 3. Update guest's qr_code_identifier
		queriedGuest.GenerateQRCodeIdentifier()
		queriedGuest.IsInvitationSent = pointer.To(true)
		if err := s.eventRepository.UpdateGuestInvitation(ctx, pointer.Get(queriedGuest)); err != nil {
			return nil
		}
		targetGuest = queriedGuest

		return nil
	})

	// 4. Update Message Template.
	toSentMessage := strings.ReplaceAll(pointer.Get(targetEvent.MessageTemplate), "{{guest_name}}", targetGuest.Name)
	toSentMessage = strings.ReplaceAll(toSentMessage, "{{event_name}}", targetEvent.Name)
	toSentMessage = strings.ReplaceAll(toSentMessage, "{{event_start_date}}", eventStartDate.Format(time.DateOnly))
	toSentMessage = strings.ReplaceAll(toSentMessage, "{{event_location}}", targetEvent.Location)
	toSentMessage = strings.ReplaceAll(toSentMessage, "{{event_digital_invitation_url}}", targetEvent.DigitalInvitationURL)
	toSentMessage = strings.ReplaceAll(toSentMessage, "{{guest_qr_code_identifier}}", targetGuest.ShortID)
	toSentMessage = strings.ReplaceAll(toSentMessage, "{{event_host}}", pointer.GetString(targetEvent.Host))

	// 5. send whatsapp message
	_, status, err = s.kirimWAClient.SendMessage(ctx, targetGuest.PhoneNumber, toSentMessage)
	fmt.Println("err", err)

	// 6. update guest invitation status
	return status, err
}

// GetGuestByShortID retrieve guest information based on it short_id.
func (s *EventService) GetGuestByShortID(ctx context.Context, guestShortID string) (guest *entity.Guest, err error) {
	return s.eventRepository.GetGuestByShortID(ctx, guestShortID)
}
