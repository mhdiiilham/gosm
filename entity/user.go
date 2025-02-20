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

	// UserRoleOrganizer represents an organizer role.
	UserRoleOrganizer UserRole = "organizer"

	// UserRoleHost represents a host role.
	UserRoleHost UserRole = "host"

	// UserRoleGuest represents a guest role.
	UserRoleGuest UserRole = "guest"
)

// User represents a user entity with personal and contact information.
type User struct {
	ID          string     `json:"id" db:"id"`
	FirstName   string     `json:"first_name" db:"first_name"`
	LastName    *string    `json:"last_name" db:"last_name"`
	Role        UserRole   `json:"role" db:"role"`
	Email       string     `json:"email" db:"email"`
	Password    string     `json:"-" db:"password"`
	CountryCode *string    `json:"country_code" db:"country_ode"`
	PhoneNumber *string    `json:"phone_number" db:"phone_umber"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"-" db:"deleted_at"`
}

// GetName returns the full name of the user.
// If the LastName field is nil, it returns only the FirstName.
func (u User) GetName() string {
	if u.LastName == nil {
		return u.FirstName
	}
	return fmt.Sprintf("%s %s", u.FirstName, *u.LastName)
}
