package delivery

// CreateEventRequest represents the payload for creating a new event.
type CreateEventRequest struct {
	Name                 string `json:"name"`
	Host                 string `json:"host"`
	Location             string `json:"location"`
	StartDate            string `json:"start_date"`
	EndDate              string `json:"end_date"`
	DigitalInvitationURL string `json:"digital_invitation_url"`
	MessageTemplate      string `json:"message_template"`
}

// AddGuestRequest represents a request to add multiple guests to an event.
type AddGuestRequest struct {
	Guests []GuestDetail `json:"guests"`
}

// GuestDetail represents the details of an individual guest.
type GuestDetail struct {
	GuestUUID   string `json:"guest_uuid"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	IsVIP       bool   `json:"is_vip"`
}

// UpdateGuestVIPStatusRequest represents a request to update single guest's vip status.
type UpdateGuestVIPStatusRequest struct {
	GuestUUID string `json:"guest_uuid"`
	IsVIP     bool   `json:"is_vip"`
}
