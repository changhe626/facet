package request

// ListEnvironmentsRequest defines query parameters for listing environments
type ListEnvironmentsRequest struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=10"`
	Keyword  string `form:"keyword"`   // Fuzzy search by name
	IsActive *bool  `form:"is_active"` // Filter by active status
}

// CreateEnvironmentRequest defines the request body for creating an environment
type CreateEnvironmentRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateEnvironmentRequest defines the request body for updating an environment
type UpdateEnvironmentRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsActive    *bool  `json:"isActive" binding:"required"`
}
