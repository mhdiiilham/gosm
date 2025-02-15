ALTER TABLE events
    DROP COLUMN host;

ALTER TABLE event_user_guests
    DROP COLUMN is_invitation_sent;