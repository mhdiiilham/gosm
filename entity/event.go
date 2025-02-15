package entity

import "time"

// Event represents an event entity with relevant metadata.
type Event struct {
	ID                   string     `json:"id"`
	UUID                 string     `json:"uuid"`
	Name                 string     `json:"name"`
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
