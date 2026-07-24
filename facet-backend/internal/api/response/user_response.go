package response

import "time"

// UserListItem defines a single user's info in the user list
type UserListItem struct {
	ID          string     `json:"id"`
	DisplayName string     `json:"displayName"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	LastLoginAt *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// ListUsersResponse defines the success response body for listing users (includes full pagination info)
type ListUsersResponse struct {
	Page       int            `json:"page"`
	PageSize   int            `json:"pageSize"`
	Total      int64          `json:"total"`
	TotalPages int            `json:"totalPages"`
	Items      []UserListItem `json:"items"`
}

// CreateUserResponse defines the success response body for creating a user
type CreateUserResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

// UserDetailResponse defines the success response body for getting user details
type UserDetailResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Role        string    `json:"role"`
	LastLoginAt time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	CreatedBy   string    `json:"createdBy"`
	UpdatedBy   string    `json:"updatedBy"`
}
