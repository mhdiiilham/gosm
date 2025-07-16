package delivery

import "time"

// CreateEventRequest represents the payload for creating a new event.
type CreateEventRequest struct {
	Title       string    `json:"name"`
	Type        string    `json:"type"`
	Location    string    `json:"location"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	Description string    `json:"description"`
	GuestCount  int       `json:"guestCount"`
}

// AddGuestRequest represents a request to add multiple guests to an event.
type AddGuestRequest struct {
	Guests []GuestDetail `json:"guests"`
}

// GuestDetail represents the details of an individual guest.
type GuestDetail struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone"`
	IsVIP       bool   `json:"vip"`
}

// UpdateGuestVIPStatusRequest represents a request to update single guest's vip status.
type UpdateGuestVIPStatusRequest struct {
	GuestUUID string `json:"guest_uuid"`
	IsVIP     bool   `json:"is_vip"`
}

// UpdateGuestAttendingAndMessage represents a request to update single guest's attending status.
type UpdateGuestAttendingAndMessage struct {
	ShortID     string `json:"short_id"`
	IsAttending bool   `json:"is_attending"`
	Message     string `json:"message"`
}

// EventResponse ...
type EventResponse struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	StartDate      string `json:"startDate"`
	EndDate        string `json:"endDate"`
	Location       string `json:"location"`
	Description    string `json:"description"`
	GuestCount     int    `json:"guestCount"`
	CheckedInCount int    `json:"checkedInCount"`
	Status         string `json:"status"`
}

type PublicAddGuestRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	IsAttending bool   `json:"isAttending"`
}
