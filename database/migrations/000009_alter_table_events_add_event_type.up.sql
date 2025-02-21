-- Create ENUM type for event types
CREATE TYPE event_type AS ENUM (
    'wedding',
    'networking',
    'conferences',
    'product_launches',
    'festival',
    'sport',
    'birthday',
    'charity',
    'cultural',
    'concert',
    'comedy',
    'gathering',
    'exhibition',
    'workshop',
    'team_building',
    'other'
);

ALTER TABLE events
    ADD COLUMN event_type event_type NOT NULL DEFAULT 'other';
