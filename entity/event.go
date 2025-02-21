package entity

import "time"

// EventType represents the type of an event.
type EventType string

var (
	// EventTypeWedding represents a wedding event.
	EventTypeWedding EventType = "wedding"

	// EventTypeNetworking represents a networking event.
	EventTypeNetworking EventType = "networking"

	// EventTypeConferences represents a conference event.
	EventTypeConferences EventType = "conferences"

	// EventTypeProductLaunches represents a product launch event.
	EventTypeProductLaunches EventType = "product_launches"

	// EventTypeFestival represents a festival event.
	EventTypeFestival EventType = "festival"

	// EventTypeSport represents a sports event.
	EventTypeSport EventType = "sport"

	// EventTypeBirthday represents a birthday event.
	EventTypeBirthday EventType = "birthday"

	// EventTypeCharity represents a charity event.
	EventTypeCharity EventType = "charity"

	// EventTypeCultural represents a cultural event.
	EventTypeCultural EventType = "cultural"

	// EventTypeConcert represents a concert event.
	EventTypeConcert EventType = "concert"

	// EventTypeComedy represents a comedy event.
	EventTypeComedy EventType = "comedy"

	// EventTypeGathering represents a social gathering event.
	EventTypeGathering EventType = "gathering"

	// EventTypeExhibitions represents an exhibition event.
	EventTypeExhibitions EventType = "exhibition"

	// EventTypeWorkshop represents a workshop event.
	EventTypeWorkshop EventType = "workshop"

	// EventTypeTeamBuilding represents a team-building event.
	EventTypeTeamBuilding EventType = "team_building"

	// EventTypeOther represents an event type that does not fit into predefined categories.
	EventTypeOther EventType = "other"
)

// ParseEventType converts a string to an EventType.
// If the input does not match any predefined event types, it returns EventTypeOther.
func ParseEventType(event string) EventType {
	switch EventType(event) {
	case EventTypeWedding,
		EventTypeNetworking,
		EventTypeConferences,
		EventTypeProductLaunches,
		EventTypeFestival,
		EventTypeSport,
		EventTypeBirthday,
		EventTypeCharity,
		EventTypeCultural,
		EventTypeConcert,
		EventTypeComedy,
		EventTypeGathering,
		EventTypeExhibitions,
		EventTypeWorkshop,
		EventTypeTeamBuilding:
		return EventType(event)
	default:
		return EventTypeOther
	}
}

// Event represents an event entity with relevant metadata.
type Event struct {
	ID                   string     `json:"id"`
	UUID                 string     `json:"uuid"`
	Name                 string     `json:"name"`
	EventType            EventType  `json:"event_type"`
	Host                 *string    `json:"host"`
	Location             string     `json:"location"`
	StartDate            string     `json:"start_date"`
	EndDate              string     `json:"end_date"`
	DigitalInvitationURL string     `json:"digital_invitation_url"`
	GuestList            []Guest    `json:"guest_list"`
	MessageTemplate      *string    `json:"message_template"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	DeletedAt            *time.Time `json:"deleted_at"`
}
