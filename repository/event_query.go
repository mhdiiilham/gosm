package repository

var (
	// SQLStatementInsertEvent inserts a new event into the "events" table.
	// It stores the event's UUID, name, location, start and end dates, and digital invitation URL.
	// The query returns the newly created event's ID.
	SQLStatementInsertEvent = `
		INSERT INTO events (
			title,
			event_type,
			description,
			location,
			start_time,
			end_time,
			created_by,
			company_id,
			guest_count
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING "id";
	`

	// SQLStatemetnUpdateEvent update event fields.
	SQLStatemetnUpdateEvent = `
		UPDATE events
			SET name = $1,
				location = $2,
				start_date = $3,
				end_date = $4,
				digital_invitation_url = $5,
				host = $6,
				message_template = $7,
				event_type = $8
		WHERE events.uuid = $9;
	`

	// SQLStatementSelectEvents retrieves a paginated list of events from the "events" table.
	// It selects events where `deleted_at` is NULL, meaning only active events are returned.
	// The results are ordered by `created_at` in descending order.
	SQLStatementSelectEvents = `
		SELECT
			events.id,
			events.title,
			events.description,
			events.location,
			events.start_time,
			events.end_time,
			events.created_by,
			users.first_name,
			events.company_id,
			companies.name,
			events.created_at,
			events.updated_at,
			events.event_type,
			events.guest_count
		FROM events
		JOIN users ON events.created_by = users.id
		JOIN companies ON events.company_id = companies.id
		WHERE events.company_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3;
	`

	// SQLStatementCountEvents counts the total number of events associated with a user.
	// It only includes events where `deleted_at` is NULL, meaning soft-deleted events are excluded.
	SQLStatementCountEvents = `
		SELECT COUNT(events.id) AS "total_events"
		FROM events
		WHERE events.company_id = $1
	`

	// SQLStatementSelectEventsByID retrieves a specific event by its UUID.
	// It ensures the event is not soft-deleted (`deleted_at IS NULL`) and belongs to the specified user.
	SQLStatementSelectEventsByID = `
		SELECT
			events.id,
			events.event_type,
			events.title,
			events.description,
			events.location,
			events.start_time,
			events.end_time,
			events.created_by,
			users.first_name,
			events.company_id,
			companies.name,
			events.created_at,
			events.updated_at,
			events.guest_count
		FROM events
		JOIN users ON events.created_by = users.id
		JOIN companies ON events.company_id = companies.id
		WHERE events.id = $1
			AND events.created_by = $2;
	`

	// SQLStatementInsertEventUserOrganizer links a user to an event as an organizer.
	SQLStatementInsertEventUserOrganizer = `
		INSERT INTO event_user_organizers (
			user_id,
			event_id
		)
		VALUES ($1, $2);
	`

	// SQLStatementAddGuestToEvent inserts a new guest into the "event_user_guests" table.
	// The guest will be associated with a specific event by `event_uuid`.
	// The query ensures that duplicate guests (same name and phone number) are not added.
	SQLStatementAddGuestToEvent = `
		INSERT INTO guests (event_id, name, email, phone, is_vip, barcode_id, is_attending)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`

	// SQLStatementGetGuestList retrieves all guests associated with a given event.
	// The results are ordered by VIP status in descending order.
	SQLStatementGetGuestList = `
		SELECT
			id,
			event_id,
			name,
			email,
			phone,
			is_vip,
			checked_in,
			barcode_id
		FROM guests
		WHERE guests.event_id = $1
		ORDER BY guests.is_vip DESC;
	`

	// SQLStatementGetGuest retrieves guests associated with a given id.
	SQLStatementGetGuest = `
		SELECT
			id,
			event_id,
			name,
			email,
			phone,
			is_vip,
			checked_in,
			barcode_id,
			is_attending
		FROM guests
		WHERE guests.barcode_id = $1
		LIMIT 1;
	`

	// SQLStatementUpdateGuestAttendAndMessage updates the guest's attendance status and message.
	SQLStatementUpdateGuestAttendAndMessage = `
		UPDATE event_user_guests
			SET
			will_attend_event = $1,
			message = $2
		WHERE event_user_guests.short_id = $3;
	`

	// SQLStatemetDeleteGuest delete guest's.
	SQLStatemetDeleteGuest = `
		DELETE FROM event_user_guests
		WHERE event_user_guests.guest_uuid = ANY($1);
	`

	// SQLStatemetSetGuestVIPStatus update guest's vip status.
	SQLStatemetSetGuestVIPStatus = `
		UPDATE event_user_guests
			SET is_vip = $1
		WHERE guest_uuid = $2;
	`

	// SQLStatementUpdateGuestInvitation update guest: is_invitation_sent, will_attend_event, and qr_code_identifier
	SQLStatementUpdateGuestInvitation = `
		UPDATE event_user_guests
			SET is_invitation_sent = $1,
				will_attend_event = $2,
				qr_code_identifier = $3
		WHERE guest_uuid = $4;
	`

	// SQLStatementUpdateGuestArrived update guest is_arrived
	SQLStatementUpdateGuestArrived = `
		UPDATE event_user_guests
			SET is_arrived = $1
		WHERE short_id = $2;
	`

	// SQLStatementUpdateGuest ...
	SQLStatementUpdateGuest = `
		UPDATE guests
			SET name = $1,
				is_attending = $2,
				phone = $3,
				message = $4
		WHERE guests.barcode_id = $5
	`

	// SQLStatementGetGuestMessages
	SQLStatementGetGuestMessages = `
		SELECT
			name,
			message
		FROM guests
		WHERE guests.event_id = $1 AND message IS NOT NULL;
	`

	// SQLStatementDeleteEvent soft delete an events.
	SQLStatementDeleteEvent = `
		UPDATE events
			SET deleted_at = now()
		WHERE events.uuid = $1;
	`

	// SQLStatementDeleteEventGuests delete guest from events.
	SQLStatementDeleteEventGuests = `
		DELETE FROM event_user_guests
		WHERE event_user_guests.event_uuid = $1;
	`
)
