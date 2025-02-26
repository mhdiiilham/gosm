package service

import (
	"context"

	"github.com/mhdiiilham/gosm/entity"
)

// CountryRepository defines the interface for country-related data operations.
type CountryRepository interface {
	GetCountries(ctx context.Context) (countries []entity.Country, err error)
}

// Country represents the service layer for country-related operations.
type Country struct {
	repository CountryRepository
}

// NewCountry creates a new Country service instance with the given repository.
func NewCountry(repository CountryRepository) *Country {
	return &Country{repository: repository}
}

// GetCountries fetches a list of countries using the repository.
func (s *Country) GetCountries(ctx context.Context) (countries []entity.Country, err error) {
	return s.repository.GetCountries(ctx)
}
