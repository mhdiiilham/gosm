package entity

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/AlekSi/pointer"
	"github.com/mhdiiilham/gosm/logger"
	"github.com/teris-io/shortid"
)

// Guest represents an event guest with their details.
type Guest struct {
	ShortID          string  `json:"short_id"`
	UUID             string  `json:"uuid"`
	Name             string  `json:"name"`
	PhoneNumber      string  `json:"phone_number"`
	IsVIP            bool    `json:"is_vip"`
	Message          *string `json:"message"`
	WillAttendEvent  *bool   `json:"will_attend_event"`
	QRCodeIdentifier *string `json:"qr_code_identifier"`
	IsInvitationSent *bool   `json:"is_invitation_sent"`
}

// GenerateQRCodeIdentifier generate a json string that encoded to base64.
func (g *Guest) GenerateQRCodeIdentifier() {
	jsonValue := map[string]any{
		"guest_uuid":        g.UUID,
		"name":              g.Name,
		"phone_number":      g.PhoneNumber,
		"will_attend_event": g.WillAttendEvent,
		"is_vip":            g.IsVIP,
		"message":           g.Message,
	}

	result, _ := json.Marshal(jsonValue)
	g.QRCodeIdentifier = pointer.ToString(base64.StdEncoding.EncodeToString(result))
}

// GetQrCodeIdentifier return string of Qr Code Identifier
func (g *Guest) GetQrCodeIdentifier() string {
	return pointer.GetString(g.QRCodeIdentifier)
}

// AssignShortID generate a short id for a guest.
// This should id use for digital invitation identifier.
func (g *Guest) AssignShortID() (err error) {
	generatedShortID, err := shortid.Generate()
	if err != nil {
		logger.Errorf(context.Background(), "Guest.AssignShortID", "failed to generate user %s short id", g.UUID)
		return UnknownError(err)
	}

	g.ShortID = generatedShortID
	return nil
}
