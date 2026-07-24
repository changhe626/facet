package response

import "time"

// Tenant defines the response body for getting a single tenant's details
type Tenant struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Website     string    `json:"website"`
	LevelID     string    `json:"levelId"`
	LevelName   string    `json:"levelName"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

// TenantListItem defines a single item in the tenant list
type TenantListItem struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	LevelID   string    `json:"levelId"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// ListTenantsResponse defines the success response body for listing tenants
type ListTenantsResponse struct {
	Page       int              `json:"page"`
	PageSize   int              `json:"pageSize"`
	Total      int64            `json:"total"`
	TotalPages int              `json:"totalPages"`
	Items      []TenantListItem `json:"items"`
}
