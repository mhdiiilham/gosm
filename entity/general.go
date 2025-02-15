package entity

// PaginationResponse represents a paginated response for API results.
type PaginationResponse struct {
	Records      any `json:"records,omitempty"`
	Page         int `json:"page,omitempty"`
	PerPage      int `json:"per_page,omitempty"`
	LastPage     int `json:"lastPage,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// PaginationRequest represents a request for paginated data retrieval.
type PaginationRequest struct {
	Page    int
	PerPage int
	Field   map[string]any
}
