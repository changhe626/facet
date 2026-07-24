package request

// ListUsersRequest defines query parameters for listing users
type ListUsersRequest struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=10"`
	Keyword  string `form:"keyword"`
}

// CreateUserRequest defines the request body for creating a user
type CreateUserRequest struct {
	Email           string `json:"email" binding:"required,email"`
	DisplayName     string `json:"displayName"`
	Role            string `json:"role" binding:"required"`
	InitialPassword string `json:"initialPassword" binding:"required,min=6"`
}

// ResetPasswordRequest defines the request body for resetting a password
type ResetPasswordRequest struct {
	// NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// UpdateUserRequest defines the request body for updating a user
type UpdateUserRequest struct {
	Email       string `json:"email" binding:"required,email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role" binding:"required"`
}
