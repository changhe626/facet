package response

import "time"

// EnvironmentListItem defines a single item in the environment list
type EnvironmentListItem struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	SDKKey    string    `json:"sdkKey"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ListEnvironmentsResponse defines the success response body for listing environments
type ListEnvironmentsResponse struct {
	Page       int                   `json:"page"`
	PageSize   int                   `json:"pageSize"`
	Total      int64                 `json:"total"`
	TotalPages int                   `json:"totalPages"`
	Items      []EnvironmentListItem `json:"items"`
}

// ActiveEnvironmentListItem defines a single item in the active environment list
type ActiveEnvironmentListItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CreateEnvironmentResponse defines the success response body for creating an environment
type CreateEnvironmentResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	SDKKey   string `json:"sdkKey"`
	IsActive bool   `json:"isActive"`
}

// UpdateEnvironmentResponse defines the success response body for updating an environment
type UpdateEnvironmentResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	SDKKey    string    `json:"sdkKey"`
	IsActive  bool      `json:"isActive"`
	UpdatedAt time.Time `json:"updatedAt"`
}
