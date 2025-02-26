package delivery

import "github.com/mhdiiilham/gosm/entity"

// GetCountriesResponse represents the response structure for a request that retrieves a list of countries.
type GetCountriesResponse struct {
	Countries []entity.Country `json:"countries"`
}
