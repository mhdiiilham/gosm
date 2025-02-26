package repository

import (
	"context"
	"database/sql"

	"github.com/mhdiiilham/gosm/entity"
	"github.com/mhdiiilham/gosm/logger"
)

// SQLStatementGetCountries is the SQL query used to retrieve the list of countries from the database.
var SQLStatementGetCountries string = `
	SELECT
		"name",
		"flag",
		"country_code",
		"phone_international_prefix"
	FROM countries
	ORDER by countries.name ASC;
`

// Country represents a repository for managing country-related database operations.
type Country struct {
	db *sql.DB
}

// NewCountry creates a new Country repository with the given database connection.
func NewCountry(db *sql.DB) *Country {
	return &Country{db: db}
}

// GetCountries retrieves a list of countries from the database.
func (r *Country) GetCountries(ctx context.Context) ([]entity.Country, error) {
	const ops = "RepositoryCountry.GetCountries"
	logger.Infof(ctx, ops, "get countries ...")

	countries := []entity.Country{}

	rows, err := r.db.QueryContext(ctx, SQLStatementGetCountries)
	if err != nil {
		logger.Errorf(ctx, ops, "failed to fetch countries from db: %v", err)
		return countries, err
	}

	for rows.Next() {
		country := entity.Country{}

		if err := rows.Scan(
			&country.Name,
			&country.Flag,
			&country.CountryCode,
			&country.PhoneInternationalPrefix,
		); err != nil {
			logger.Errorf(ctx, ops, "failed to scan country from db: %v", err)
		}

		countries = append(countries, country)
	}

	return countries, nil
}
