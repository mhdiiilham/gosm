package delivery

import "github.com/mhdiiilham/gosm/entity"

// SignUpRequest represents the payload required for a user to sign up.
type SignUpRequest struct {
	FirstName   string          `json:"first_name"`
	LastName    string          `json:"last_name"`
	Email       string          `json:"email"`
	PhoneNumber string          `json:"phone_number"`
	Password    string          `json:"password"`
	Role        entity.UserRole `json:"role"`
}

// SignInRequest represents the payload required for a user to sign in.
type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

// AccessTokenResponse represents the response returned after successful authentication.
type AccessTokenResponse struct {
	Email       string `json:"email"`
	AccessToken string `json:"access_token"`
	ExpiresAt   string `json:"expires_at"`
}
