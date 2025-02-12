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
	ID          string     `db:"id"`
	FirstName   string     `db:"first_name"`
	LastName    *string    `db:"last_name"`
	Role        UserRole   `db:"role"`
	Email       string     `db:"email"`
	Password    string     `db:"password"`
	CountryCode *string    `db:"country_ode"`
	PhoneNumber *string    `db:"phone_umber"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

// GetName returns the full name of the user.
// If the LastName field is nil, it returns only the FirstName.
func (u User) GetName() string {
	if u.LastName == nil {
		return u.FirstName
	}
	return fmt.Sprintf("%s %s", u.FirstName, *u.LastName)
}
