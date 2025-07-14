package entity

import (
	"fmt"
	"time"
)

// UserRole represents different roles a user can have in the system.
type UserRole string

var (
	// UserRoleSuperAdmin represents an super admin role.
	UserRoleSuperAdmin UserRole = "super_admin"

	// UserRoleEOOrganizer represents an eo organizer role.
	UserRoleEOOrganizer UserRole = "eo_admin"

	// UserRoleCrew represents an crew role.
	UserRoleCrew UserRole = "crew"

	// UserRoleHost represents an independent host role.
	UserRoleHost UserRole = "independent_host"

	// UserRoleGuest represents a guest role.
	UserRoleGuest UserRole = "guest"
)

// User represents a user entity with personal and contact information.
type User struct {
	ID          int        `json:"id"`
	FirstName   string     `json:"first_name"`
	LastName    *string    `json:"last_name"`
	Role        UserRole   `json:"role"`
	Email       string     `json:"email"`
	Password    string     `json:"-"`
	PhoneNumber *string    `json:"phone_number"`
	JobTitle    *string    `json:"job_title"`
	CompanyID   *int       `json:"company_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"-"`
}

// GetName returns the full name of the user.
// If the LastName field is nil, it returns only the FirstName.
func (u User) GetName() string {
	if u.LastName == nil {
		return u.FirstName
	}
	return fmt.Sprintf("%s %s", u.FirstName, *u.LastName)
}

// GetCompanyID ...
func (u User) GetCompanyID() int {
	if u.CompanyID == nil {
		return 0
	}

	return *u.CompanyID
}
