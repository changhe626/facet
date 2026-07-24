package middleware

import (
	"facet-backend/internal/api/response"
	"facet-backend/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole checks that the authenticated user has the exact specified role.
func RequireRole(role domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(string(UserRoleKey))
		if !exists {
			response.CustomError(c, http.StatusForbidden, "access denied")
			c.Abort()
			return
		}
		if userRole.(string) != string(role) {
			response.CustomError(c, http.StatusForbidden, "access denied")
			c.Abort()
			return
		}
		c.Next()
	}
}

var roleWeights = map[domain.UserRole]int{
	domain.RoleSystemAdmin:      100,
	domain.RoleTenantSuperAdmin: 80,
	domain.RoleTenantAdmin:      60,
	domain.RoleTenantViewer:     40,
}

// RequireMinRole checks that the authenticated user has at least the specified role level.
// Role hierarchy: system_admin > tenant_super_admin > tenant_admin > tenant_viewer
func RequireMinRole(required domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(string(UserRoleKey))
		if !exists {
			response.CustomError(c, http.StatusForbidden, "access denied")
			c.Abort()
			return
		}
		userWeight, ok := roleWeights[domain.UserRole(userRole.(string))]
		if !ok {
			response.CustomError(c, http.StatusForbidden, "access denied")
			c.Abort()
			return
		}
		requiredWeight, ok := roleWeights[required]
		if !ok {
			response.CustomError(c, http.StatusForbidden, "access denied")
			c.Abort()
			return
		}
		if userWeight < requiredWeight {
			response.CustomError(c, http.StatusForbidden, "access denied")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireTenantWriteRole checks that the authenticated user has at least tenant_admin role for write operations.
// Read operations (GET) are allowed for tenant_viewer and above.
// Write operations (POST, PUT, DELETE) require tenant_admin or above.
func RequireTenantWriteRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(string(UserRoleKey))
		if !exists {
			response.CustomError(c, http.StatusForbidden, "access denied")
			c.Abort()
			return
		}
		role := domain.UserRole(userRole.(string))

		// system_admin can do everything
		if role == domain.RoleSystemAdmin {
			c.Next()
			return
		}

		// Write operations require at least tenant_admin
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" || c.Request.Method == "PATCH" {
			if role != domain.RoleTenantSuperAdmin && role != domain.RoleTenantAdmin {
				response.CustomError(c, http.StatusForbidden, "access denied: insufficient privileges")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}