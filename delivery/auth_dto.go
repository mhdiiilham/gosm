package delivery

import (
	"github.com/AlekSi/pointer"
	"github.com/mhdiiilham/gosm/entity"
)

// SignUpRequest represents the payload required for a user to sign up.
type SignUpRequest struct {
	FirstName   string          `json:"first_name"`
	LastName    string          `json:"last_name"`
	Email       string          `json:"email"`
	PhoneNumber string          `json:"phone_number"`
	Password    string          `json:"password"`
	CompanyName string          `json:"company_name"`
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
	AccessToken string           `json:"access_token"`
	User        UserResponse     `json:"user"`
	Company     *CompanyResponse `json:"company"`
}

// UserResponse ...
type UserResponse struct {
	ID       int             `json:"id"`
	Name     string          `json:"name"`
	Email    string          `json:"email"`
	Phone    *string         `json:"phone"`
	JobTitle *string         `json:"job_title"`
	Address  *string         `json:"address"`
	Role     entity.UserRole `json:"role"`
}

// CompanyResponse ...
type CompanyResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Website     string `json:"website"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

func CompanyResponseFromEntity(company entity.Company) CompanyResponse {
	return CompanyResponse{
		ID:          company.ID,
		Name:        company.Name,
		Email:       company.Email,
		Phone:       company.Phone,
		Website:     pointer.Get(company.Website),
		Address:     pointer.Get(company.Address),
		Description: pointer.Get(company.Description),
	}
}

// const user = ref({
//   name: 'John Doe',
//   email: 'john.doe@example.com',
//   phone: '(555) 123-4567',
//   jobTitle: 'Event Manager',
//   address: '123 Main St, Anytown, CA 12345'
// });

type ProfileResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	JobTitle string `json:"jobTitle"`
	Role     string `json:"role"`
}

func ProfileResponseFromEntity(user *entity.User) ProfileResponse {
	if user == nil {
		return ProfileResponse{}
	}

	return ProfileResponse{
		ID:       user.ID,
		Name:     user.GetName(),
		Email:    user.Email,
		Phone:    pointer.Get(user.PhoneNumber),
		JobTitle: pointer.Get(user.JobTitle),
		Role:     string(user.Role),
	}
}
