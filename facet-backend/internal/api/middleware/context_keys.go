package middleware

// contextKey is a private type used to create unique keys in context, avoiding conflicts with other packages
type contextKey string

// Defines keys for storing and retrieving information in the Gin context
const (
	// UserIDKey stores the ID of the currently logged-in user
	UserIDKey contextKey = "userID"

	// TenantIDKey stores the tenant ID of the currently logged-in user
	TenantIDKey contextKey = "tenantID"

	// UserRoleKey stores the role of the currently logged-in user
	UserRoleKey contextKey = "userRole"
)
