package entity

// AuthResponse represents the response body for successful authentication.
type AuthResponse struct {
	AccessToken string   `json:"access_token"`
	ExpiresAt   string   `json:"expires_at"`
	Email       string   `json:"email"`
	Role        UserRole `json:"role"`
}
