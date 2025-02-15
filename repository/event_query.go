package repository

var (
	// SQLStatementInsertEvent inserts a new event into the "events" table.
	// It stores the event's UUID, name, location, start and end dates, and digital invitation URL.
	// The query returns the newly created event's ID.
	SQLStatementInsertEvent = `
		INSERT INTO events (
			uuid,
			name,
			location,
			start_date,
			end_date,
			digital_invitation_url,
			host,
			message_template
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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
				message_template = $7
		WHERE events.uuid = $8;
	`

	// SQLStatementSelectEvents retrieves a paginated list of events from the "events" table.
	// It selects events where `deleted_at` is NULL, meaning only active events are returned.
	// The results are ordered by `created_at` in descending order.
	SQLStatementSelectEvents = `
		SELECT
			id,
			uuid,
			name,
			location,
			start_date,
			end_date,
			digital_invitation_url,
			created_at,
			updated_at,
			host
		FROM events
		JOIN event_user_organizers ON events.id = event_user_organizers.event_id
		WHERE
			deleted_at IS NULL
			AND event_user_organizers.user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3;
	`

	// SQLStatementCountEvents counts the total number of events associated with a user.
	// It only includes events where `deleted_at` is NULL, meaning soft-deleted events are excluded.
	SQLStatementCountEvents = `
		SELECT COUNT(events.id) AS "total_events"
		FROM events
		JOIN event_user_organizers ON events.id = event_user_organizers.event_id
		WHERE
			deleted_at IS NULL
			AND event_user_organizers.user_id = $1;
	`

	// SQLStatementSelectEventsByUUID retrieves a specific event by its UUID.
	// It ensures the event is not soft-deleted (`deleted_at IS NULL`) and belongs to the specified user.
	SQLStatementSelectEventsByUUID = `
		SELECT
			id,
			uuid,
			name,
			location,
			start_date,
			end_date,
			digital_invitation_url,
			created_at,
			updated_at,
			host,
			message_template
		FROM events
		JOIN event_user_organizers ON events.id = event_user_organizers.event_id
		WHERE deleted_at IS NULL
			AND event_user_organizers.user_id = $1
			AND events.uuid = $2;
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
		INSERT INTO event_user_guests (event_uuid, guest_uuid, name, phone_number, is_vip, short_id)
		SELECT 
			$1,
    		$2,
    		$3,
    		$4,
			$5,
			$6
		WHERE NOT EXISTS (
    		SELECT 1 FROM event_user_guests
    		WHERE name = $7
      		AND phone_number = $8
		);
	`

	// SQLStatementGetGuestList retrieves all guests associated with a given event.
	// The results are ordered by VIP status in descending order.
	SQLStatementGetGuestList = `
		SELECT
			guest_uuid,
			name,
			phone_number,
			message,
			will_attend_event,
			qr_code_identifier,
			is_vip::BOOLEAN,
			is_invitation_sent,
			short_id
		FROM event_user_guests
		WHERE event_uuid = $1
		ORDER BY event_user_guests.is_vip DESC NULLS LAST;
	`

	// SQLStatementGetGuest retrieves guests associated with a given id.
	SQLStatementGetGuest = `
		SELECT
			guest_uuid,
			name,
			phone_number,
			message,
			will_attend_event,
			qr_code_identifier,
			is_vip::BOOLEAN,
			is_invitation_sent,
			short_id
		FROM event_user_guests
		WHERE guest_uuid = $1
		ORDER BY event_user_guests.is_vip DESC NULLS LAST;
	`

	// SQLStatementGetGuestByShortID retrieves guest with give short_id.
	SQLStatementGetGuestByShortID = `
		SELECT
			guest_uuid,
			name,
			phone_number,
			message,
			will_attend_event,
			qr_code_identifier,
			is_vip::BOOLEAN,
			is_invitation_sent,
			short_id
		FROM event_user_guests
		WHERE short_id = $1
		ORDER BY event_user_guests.is_vip DESC NULLS LAST;
	`

	// SQLStatementUpdateGuestAttendAndMessage updates the guest's attendance status and message.
	SQLStatementUpdateGuestAttendAndMessage = `
		UPDATE event_user_guests
			SET
			will_attend_event = $1,
			message = $2
		WHERE event_user_guests.guest_uuid = $3;
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
)
