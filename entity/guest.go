package entity

// Guest represents an event guest with their details.
type Guest struct {
	ID          int    `json:"id"`
	EventID     int    `json:"eventId"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	IsVIP       bool   `json:"vip"`
	CheckedIn   bool   `json:"checkedIn"`
	CheckedInBy int    `json:"checkedInBy"`
	BarcodeID   string `json:"barcode"`
}
