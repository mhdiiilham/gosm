package entity

// Country represents a country with its relevant details.
type Country struct {
	Name                     string `json:"name"`
	Flag                     string `json:"flag"`
	CountryCode              string `json:"country_code"`
	PhoneInternationalPrefix int    `json:"phone_international_prefix"`
}
