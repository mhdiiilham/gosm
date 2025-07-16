package repository

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/lib/pq"
	"github.com/mhdiiilham/gosm/entity"
	"github.com/mhdiiilham/gosm/logger"
	"github.com/mhdiiilham/gosm/pkg"
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
func (r *EventRepository) CreateEvent(ctx context.Context, event entity.Event) (createdEvent *entity.Event, err error) {
	row := r.db.QueryRowContext(
		ctx,
		SQLStatementInsertEvent,
		event.Title,
		event.Type,
		event.Description,
		event.Location,
		event.StartDate,
		event.EndDate,
		event.CreatedBy.ID,
		event.Company.ID,
		event.GuestCount,
	)

	if err := row.Scan(&event.ID); err != nil {
		return nil, err
	}

	return &event, nil
}

// GetEvent retrieves an event by its UUID for a specific user.
func (r *EventRepository) GetEvent(ctx context.Context, tx *sql.Tx, userID, eventID int) (event *entity.Event, err error) {
	row := tx.QueryRowContext(ctx, SQLStatementSelectEventsByID, eventID, userID)

	event = &entity.Event{}
	if err := row.Scan(
		&event.ID,
		&event.Type,
		&event.Title,
		&event.Description,
		&event.Location,
		&event.StartDate,
		&event.EndDate,
		&event.CreatedBy.ID,
		&event.CreatedBy.Name,
		&event.Company.ID,
		&event.Company.Name,
		&event.CreatedAt,
		&event.UpdatedAt,
		&event.GuestCount,
	); err != nil {
		return nil, err
	}

	return event, nil
}

// GetEvents retrieves a paginated list of events for a specific company.
func (r *EventRepository) GetEvents(ctx context.Context, companyID int, limit, offset int) ([]entity.Event, int, error) {
	const ops = "EventRepository.GetEvents"

	var totalEvents int
	if err := r.db.QueryRowContext(ctx, SQLStatementCountEvents, companyID).Scan(&totalEvents); err != nil {
		logger.Errorf(ctx, ops, "failed to fetch events: %v", err)
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, SQLStatementSelectEvents, companyID, limit, offset)
	if err != nil {
		logger.Errorf(ctx, ops, "failed to fetch events: %v", err)
		return nil, 0, err
	}

	events := []entity.Event{}
	for rows.Next() {
		event := entity.Event{}
		var eventType string
		if err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.Location,
			&event.StartDate,
			&event.EndDate,
			&event.CreatedBy.ID,
			&event.CreatedBy.Name,
			&event.Company.ID,
			&event.Company.Name,
			&event.CreatedAt,
			&event.UpdatedAt,
			&eventType,
			&event.GuestCount,
		); err != nil {
			logger.Errorf(ctx, ops, "failed to scan an event: %v", err)
		}

		event.Type = entity.ParseEventType(eventType)

		events = append(events, event)

	}

	return events, totalEvents, nil
}

// AddGuests adds a list of guests to an event.
func (r *EventRepository) AddGuests(ctx context.Context, eventID int, guestList []entity.Guest) (numberOfSuccess int, err error) {
	const ops = "EventRepository.AddGuests"

	for _, guest := range guestList {
		barcodeID := guest.BarcodeID
		if barcodeID == "" {
			barcodeID, _ = pkg.GeneratePumBookID(strconv.Itoa(eventID))
		}

		r, err := r.db.ExecContext(
			ctx, SQLStatementAddGuestToEvent,
			eventID,
			guest.Name,
			guest.Email,
			guest.Phone,
			guest.IsVIP,
			barcodeID,
			guest.IsAttending,
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
func (r *EventRepository) GetGuests(ctx context.Context, eventID int) (response []entity.Guest, err error) {
	response = []entity.Guest{}

	rows, err := r.db.QueryContext(ctx, SQLStatementGetGuestList, eventID)
	if err != nil {
		logger.Errorf(ctx, "EventRepository.GetGuests", "failed to retrieve list of guest: %v", err)
		return nil, err
	}

	for rows.Next() {
		guest := entity.Guest{}
		rows.Scan(
			&guest.ID,
			&guest.EventID,
			&guest.Name,
			&guest.Email,
			&guest.Phone,
			&guest.IsVIP,
			&guest.CheckedIn,
			&guest.BarcodeID,
		)

		response = append(response, guest)

	}

	return response, nil
}

// DeleteGuests delete list of selected guest.
func (r *EventRepository) DeleteGuests(ctx context.Context, userID int, guestIDs []int) error {
	toDeleteIDs := pq.StringArray{}
	for _, g := range guestIDs {
		toDeleteIDs = append(toDeleteIDs, strconv.Itoa(g))
	}

	if _, err := r.db.ExecContext(ctx, SQLStatemetDeleteGuest, toDeleteIDs); err != nil {
		logger.Errorf(ctx, "EventRepository.GetGuests", "failed to delete guest: %v", err)
		return nil
	}

	return nil
}

// UpdateGuestVIPStatus update the guest's vip status.
func (r *EventRepository) UpdateGuestVIPStatus(ctx context.Context, guestID int, vipStatus bool) error {
	isVIPStatus := "0"
	if vipStatus {
		isVIPStatus = "1"
	}

	if _, err := r.db.ExecContext(ctx, SQLStatemetSetGuestVIPStatus, isVIPStatus, guestID); err != nil {
		logger.Errorf(ctx, "EventRepository.GetGuests", "failed to update guest vip status: %v", err)
		return nil
	}

	return nil
}

// GetGuest get a guest of an event.
func (r *EventRepository) GetGuest(ctx context.Context, barcodeID string) (guest *entity.Guest, err error) {
	targetGuest := entity.Guest{}
	if err := r.db.QueryRowContext(ctx, SQLStatementGetGuest, barcodeID).Scan(
		&targetGuest.ID,
		&targetGuest.EventID,
		&targetGuest.Name,
		&targetGuest.Email,
		&targetGuest.Phone,
		&targetGuest.IsVIP,
		&targetGuest.CheckedIn,
		&targetGuest.BarcodeID,
		&targetGuest.IsAttending,
	); err != nil {
		logger.Errorf(ctx, "EventRepository.GetGuest", "failed to retrieve guest: %v", err)
		return nil, err
	}

	return &targetGuest, nil
}

// UpdateEvent update an existing event.
func (r *EventRepository) UpdateEvent(ctx context.Context, event entity.Event) (err error) {
	// _, err = r.db.ExecContext(
	// 	ctx,
	// 	SQLStatemetnUpdateEvent,
	// 	event.Name,
	// 	event.Location,
	// 	event.StartDate,
	// 	event.EndDate,
	// 	event.DigitalInvitationURL,
	// 	event.Host,
	// 	event.MessageTemplate,
	// 	event.EventType,
	// 	event.UUID,
	// )
	// if err != nil {
	// 	logger.Errorf(ctx, "EventRepository.UpdateEvent", "failed to update event: %v", err)
	// 	return nil
	// }

	return nil
}

// UpdateGuestInvitation update given guest_uuid invitation related values.
func (r *EventRepository) UpdateGuestInvitation(ctx context.Context, guest entity.Guest) (err error) {
	// willAttendEvent := "0"

	// if guest.WillAttendEvent != nil {
	// 	if pointer.Get(guest.WillAttendEvent) {
	// 		willAttendEvent = "1"
	// 	}
	// }

	// if _, err := r.db.ExecContext(ctx, SQLStatementUpdateGuestInvitation,
	// 	guest.IsInvitationSent,
	// 	willAttendEvent,
	// 	guest.GetQrCodeIdentifier(),
	// 	guest.UUID,
	// ); err != nil {
	// 	logger.Errorf(ctx, "EventRepository.UpdateGuestInvitation", "failed to update guest's invitation related fields: %v", err)
	// 	return err
	// }

	return nil
}

// UpdateGuestAttendingStatus update guest's attending status
func (r *EventRepository) UpdateGuestAttendingStatus(ctx context.Context, guestID int, isAttending bool, message string) (err error) {
	willAttend := "0"
	if isAttending {
		willAttend = "1"
	}

	if _, err := r.db.ExecContext(ctx, SQLStatementUpdateGuestAttendAndMessage, willAttend, message, guestID); err != nil {
		logger.Errorf(ctx, "EventRepository.UpdateGuestAttendingStatus", "failed to update guest: %v", err)
	}

	return nil
}

// DeleteEvent soft delete an event based on it given uuid.
func (r *EventRepository) DeleteEvent(ctx context.Context, eventID int) (bool, error) {
	result, err := r.db.ExecContext(ctx, SQLStatementDeleteEvent, eventID)
	if err != nil {
		logger.Errorf(ctx, "EventRepository.DeleteEvent", "failed to delete event: %v", err)
		return false, err
	}

	// ignore result and error.
	_, err = r.db.ExecContext(ctx, SQLStatementDeleteEventGuests, eventID)
	if err != nil {
		logger.Errorf(ctx, "EventRepository.DeleteEvent", "failed to remove guests from event: %v", err)
	}

	rowAffected, _ := result.RowsAffected()
	return int(rowAffected) != 0, nil
}

// SetGuestIsArrived update an guest `is_arrived`
func (r *EventRepository) SetGuestIsArrived(ctx context.Context, guestID int, isArrived bool) (err error) {
	_, err = r.db.ExecContext(ctx, SQLStatementUpdateGuestArrived, isArrived, guestID)
	if err != nil {
		logger.Errorf(ctx, "EventRepository.DeleteEvent", "failed to update guest is_arrived: %v", err)
	}

	return nil
}

func (r *EventRepository) UpdateGuest(ctx context.Context, guestID, name, phone string, isAttending bool) error {
	_, err := r.db.ExecContext(ctx, SQLStatementUpdateGuest, name, isAttending, phone, guestID)
	if err != nil {
		logger.Errorf(ctx, "EventRepository.DeleteEvent", "failed to update guest: %v", err)
	}

	return nil
}
