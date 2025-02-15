ALTER TABLE event_user_guests
    ADD COLUMN short_id VARCHAR NOT NULL UNIQUE;