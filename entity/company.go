package entity

import (
	"time"
)

// Company represents a company entity.
type Company struct {
	ID          int
	Name        string
	Email       string
	Phone       string
	Address     *string
	Website     *string
	Description *string
	LogoURL     *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
