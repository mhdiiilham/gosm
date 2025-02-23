package delivery

// GetCountriesResponse represents the response structure for a request that retrieves a list of countries.
type GetCountriesResponse struct {
	Countries []Country `json:"countries"`
}

// Country represents a country with its relevant details.
type Country struct {
	Name                     string `json:"name"`
	Flag                     string `json:"flag"`
	CountryCode              string `json:"country_code"`
	PhoneInternationalPrefix int    `json:"phone_international_prefix"`
}
