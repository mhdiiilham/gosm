package repository

import (
	"context"
	"database/sql"

	"github.com/AlekSi/pointer"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/mhdiiilham/gosm/entity"
	"github.com/mhdiiilham/gosm/logger"
)

// EventRepository provides methods for interacting with the "events" database table.
type EventRepository struct {
	db *sql.DB
}

// NewEventRepository initializes a new EventRepository with a given database connection.
func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

// RunInTransactions executes a function within a database transaction.
func (r *EventRepository) RunInTransactions(ctx context.Context, fn entity.TransactionFunc) error {
	const ops = "EventRepository.RunInTransactions"
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Errorf(ctx, ops, "failed to begin database transaction: %v", err)
		return err
	}

	if err := fn(ctx, tx); err != nil {
		tx.Rollback()
		logger.Errorf(ctx, ops, "failed to execute transaction: %v", err)
		return err
	}

	return tx.Commit()
}

// CreateEvent inserts a new event into the "events" table and returns the created event.
// It assigns a generated event ID to the input entity.
func (r *EventRepository) CreateEvent(ctx context.Context, userID string, event entity.Event) (createdEvent *entity.Event, err error) {
	row := r.db.QueryRowContext(
		ctx,
		SQLStatementInsertEvent,
		event.UUID,
		event.Name,
		event.Location,
		event.StartDate,
		event.EndDate,
		event.DigitalInvitationURL,
		event.Host,
		event.MessageTemplate,
		event.EventType,
	)

	if err := row.Scan(&event.ID); err != nil {
		return nil, err
	}

	_, err = r.db.ExecContext(ctx, SQLStatementInsertEventUserOrganizer, userID, event.ID)
	if err != nil {
		logger.Errorf(ctx, "EventRepository.CreateEvent", "failed to insert user id as organizer")
		// handle error later
	}

	return &event, nil
}

// GetEvent retrieves an event by its UUID for a specific user.
func (r *EventRepository) GetEvent(ctx context.Context, tx *sql.Tx, userID, UUID string) (event *entity.Event, err error) {
	row := tx.QueryRowContext(ctx, SQLStatementSelectEventsByUUID, userID, UUID)

	event = &entity.Event{}
	if err := row.Scan(
		&event.ID,
		&event.UUID,
		&event.Name,
		&event.Location,
		&event.StartDate,
		&event.EndDate,
		&event.DigitalInvitationURL,
		&event.CreatedAt,
		&event.UpdatedAt,
		&event.Host,
		&event.MessageTemplate,
		&event.EventType,
	); err != nil {
		return nil, err
	}

	return event, nil
}

// GetEvents retrieves a paginated list of events for a specific user.
func (r *EventRepository) GetEvents(ctx context.Context, userID string, limit, offset int) ([]entity.Event, int, error) {
	const ops = "EventRepository.GetEvents"

	var totalEvents int
	if err := r.db.QueryRowContext(ctx, SQLStatementCountEvents, userID).Scan(&totalEvents); err != nil {
		logger.Errorf(ctx, ops, "failed to fetch events: %v", err)
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, SQLStatementSelectEvents, userID, limit, offset)
	if err != nil {
		logger.Errorf(ctx, ops, "failed to fetch events: %v", err)
		return nil, 0, err
	}

	var events []entity.Event
	for rows.Next() {
		event := entity.Event{}
		if err := rows.Scan(
			&event.ID,
			&event.UUID,
			&event.Name,
			&event.Location,
			&event.StartDate,
			&event.EndDate,
			&event.DigitalInvitationURL,
			&event.CreatedAt,
			&event.UpdatedAt,
			&event.Host,
		); err != nil {
			logger.Errorf(ctx, ops, "failed to scan an event: %v", err)
		}

		events = append(events, event)

	}

	return events, totalEvents, nil
}

// AddGuests adds a list of guests to an event.
func (r *EventRepository) AddGuests(ctx context.Context, eventUUID string, guestList []entity.Guest) (numberOfSuccess int, err error) {
	const ops = "EventRepository.AddGuests"

	for _, guest := range guestList {
		generatedUUID := uuid.NewString()
		isVIP := "0"
		if guest.IsVIP {
			isVIP = "1"
		}

		r, err := r.db.ExecContext(
			ctx, SQLStatementAddGuestToEvent,
			eventUUID,
			generatedUUID,
			guest.Name,
			guest.PhoneNumber,
			isVIP,
			guest.ShortID,
			guest.Name,
			guest.PhoneNumber,
		)
		if err != nil {
			logger.Errorf(ctx, ops, "failed to add guest to an event: %v", err)
		}

		rowsAffected, _ := r.RowsAffected()
		if rowsAffected == 1 {
			numberOfSuccess++
		}
	}

	return numberOfSuccess, nil
}

// GetGuests get a list of guest of an event.
// TODO: need to apply worker pattern here.
//
//	Will have to query guests per batch.
//	each probably like 100?
//	Also have to limit the number of workers (like 2 or 3?)
func (r *EventRepository) GetGuests(ctx context.Context, tx *sql.Tx, eventUUID string) (response []entity.Guest, err error) {
	response = []entity.Guest{}

	rows, err := tx.QueryContext(ctx, SQLStatementGetGuestList, eventUUID)
	if err != nil {
		logger.Errorf(ctx, "EventRepository.GetGuests", "failed to retrieve list of guest: %v", err)
		return nil, err
	}

	for rows.Next() {
		guest := entity.Guest{}
		rows.Scan(
			&guest.UUID,
			&guest.Name,
			&guest.PhoneNumber,
			&guest.Message,
			&guest.WillAttendEvent,
			&guest.QRCodeIdentifier,
			&guest.IsVIP,
			&guest.IsInvitationSent,
			&guest.ShortID,
		)

		response = append(response, guest)

	}

	return response, nil
}

// DeleteGuests delete list of selected guest.
func (r *EventRepository) DeleteGuests(ctx context.Context, userID string, guestUUIDs []string) error {
	toDeleteIDs := pq.StringArray{}
	for _, g := range guestUUIDs {
		toDeleteIDs = append(toDeleteIDs, g)
	}

	if _, err := r.db.ExecContext(ctx, SQLStatemetDeleteGuest, toDeleteIDs); err != nil {
		logger.Errorf(ctx, "EventRepository.GetGuests", "failed to delete guest: %v", err)
		return nil
	}

	return nil
}

// UpdateGuestVIPStatus update the guest's vip status.
func (r *EventRepository) UpdateGuestVIPStatus(ctx context.Context, guestUUID string, vipStatus bool) error {
	isVIPStatus := "0"
	if vipStatus {
		isVIPStatus = "1"
	}

	if _, err := r.db.ExecContext(ctx, SQLStatemetSetGuestVIPStatus, isVIPStatus, guestUUID); err != nil {
		logger.Errorf(ctx, "EventRepository.GetGuests", "failed to update guest vip status: %v", err)
		return nil
	}

	return nil
}

// GetGuest get a guest of an event.
func (r *EventRepository) GetGuest(ctx context.Context, guestUUID string) (guest *entity.Guest, err error) {
	targetGuest := entity.Guest{}
	if err := r.db.QueryRowContext(ctx, SQLStatementGetGuest, guestUUID).Scan(
		&targetGuest.UUID,
		&targetGuest.Name,
		&targetGuest.PhoneNumber,
		&targetGuest.Message,
		&targetGuest.WillAttendEvent,
		&targetGuest.QRCodeIdentifier,
		&targetGuest.IsVIP,
		&targetGuest.IsInvitationSent,
		&targetGuest.ShortID,
	); err != nil {
		logger.Errorf(ctx, "EventRepository.GetGuest", "failed to retrieve guest: %v", err)
		// handle error later
		return nil, err
	}

	return &targetGuest, nil
}

// UpdateEvent update an existing event.
func (r *EventRepository) UpdateEvent(ctx context.Context, event entity.Event) (err error) {
	_, err = r.db.ExecContext(
		ctx,
		SQLStatemetnUpdateEvent,
		event.Name,
		event.Location,
		event.StartDate,
		event.EndDate,
		event.DigitalInvitationURL,
		event.Host,
		event.MessageTemplate,
		event.EventType,
		event.UUID,
	)
	if err != nil {
		logger.Errorf(ctx, "EventRepository.UpdateEvent", "failed to update event: %v", err)
		return nil
	}

	return nil
}

// UpdateGuestInvitation update given guest_uuid invitation related values.
func (r *EventRepository) UpdateGuestInvitation(ctx context.Context, guest entity.Guest) (err error) {
	willAttendEvent := "0"

	if guest.WillAttendEvent != nil {
		if pointer.Get(guest.WillAttendEvent) {
			willAttendEvent = "1"
		}
	}

	if _, err := r.db.ExecContext(ctx, SQLStatementUpdateGuestInvitation,
		guest.IsInvitationSent,
		willAttendEvent,
		guest.GetQrCodeIdentifier(),
		guest.UUID,
	); err != nil {
		logger.Errorf(ctx, "EventRepository.UpdateGuestInvitation", "failed to update guest's invitation related fields: %v", err)
		return err
	}

	return nil
}

// GetGuestByShortID get a guest of an event using it short_id.
func (r *EventRepository) GetGuestByShortID(ctx context.Context, guestShortID string) (guest *entity.Guest, err error) {
	targetGuest := entity.Guest{}
	if err := r.db.QueryRowContext(ctx, SQLStatementGetGuestByShortID, guestShortID).Scan(
		&targetGuest.UUID,
		&targetGuest.Name,
		&targetGuest.PhoneNumber,
		&targetGuest.Message,
		&targetGuest.WillAttendEvent,
		&targetGuest.QRCodeIdentifier,
		&targetGuest.IsVIP,
		&targetGuest.IsInvitationSent,
		&targetGuest.ShortID,
	); err != nil {
		logger.Errorf(ctx, "EventRepository.GetGuestByShortID", "failed to retrieve guest: %v", err)
		// handle error later
		return nil, err
	}

	return &targetGuest, nil
}

// UpdateGuestAttendingStatus update guest's attending status
func (r *EventRepository) UpdateGuestAttendingStatus(ctx context.Context, guestShortID string, isAttending bool, message string) (err error) {
	willAttend := "0"
	if isAttending {
		willAttend = "1"
	}

	if _, err := r.db.ExecContext(ctx, SQLStatementUpdateGuestAttendAndMessage, willAttend, message, guestShortID); err != nil {
		logger.Errorf(ctx, "EventRepository.UpdateGuestAttendingStatus", "failed to update guest: %v", err)
	}

	return nil
}

// DeleteEvent soft delete an event based on it given uuid.
func (r *EventRepository) DeleteEvent(ctx context.Context, eventUUID string) (bool, error) {
	result, err := r.db.ExecContext(ctx, SQLStatementDeleteEvent, eventUUID)
	if err != nil {
		logger.Errorf(ctx, "EventRepository.DeleteEvent", "failed to delete event: %v", err)
		return false, err
	}

	// ignore result and error.
	_, err = r.db.ExecContext(ctx, SQLStatementDeleteEventGuests, eventUUID)
	if err != nil {
		logger.Errorf(ctx, "EventRepository.DeleteEvent", "failed to remove guests from event: %v", err)
	}

	rowAffected, _ := result.RowsAffected()
	return int(rowAffected) != 0, nil
}

// SetGuestIsArrived update an guest `is_arrived`
func (r *EventRepository) SetGuestIsArrived(ctx context.Context, guestShortID string, isArrived bool) (err error) {
	_, err = r.db.ExecContext(ctx, SQLStatementUpdateGuestArrived, isArrived, guestShortID)
	if err != nil {
		logger.Errorf(ctx, "EventRepository.DeleteEvent", "failed to update guest is_arrived: %v", err)
	}

	return nil
}
