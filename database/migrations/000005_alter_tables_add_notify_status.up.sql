ALTER TABLE events
    ADD COLUMN host VARCHAR;

ALTER TABLE event_user_guests
    ADD COLUMN is_invitation_sent BOOLEAN;