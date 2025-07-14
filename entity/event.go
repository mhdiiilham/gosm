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
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Type        EventType `json:"type"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	CreatedBy   IDName    `json:"createdBy"`
	Company     IDName    `json:"company"`
	GuestCount  int       `json:"guestCount"`
	Guests      []Guest   `json:"guests"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
