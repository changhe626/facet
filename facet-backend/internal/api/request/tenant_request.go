package request

// CreateTenantInfo defines the tenant's own information when creating a tenant
type CreateTenantInfo struct {
	Name        string `json:"name" binding:"required"`
	LevelID     string `json:"levelId"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Website     string `json:"website"`
}

// CreateAdminUserInfo defines the super admin's information when creating a tenant
type CreateAdminUserInfo struct {
	Role                 string `json:"role" binding:"required"`
	AdminEmail           string `json:"adminEmail" binding:"required,email"`
	AdminDisplayName     string `json:"adminDisplayName"`
	AdminInitialPassword string `json:"adminInitialPassword" binding:"required,min=8"`
}

// CreateTenantRequest defines the complete request body for creating a tenant
type CreateTenantRequest struct {
	TenantInfo CreateTenantInfo    `json:"tenant_info" binding:"required"`
	UserInfo   CreateAdminUserInfo `json:"user_info" binding:"required"`
}

// UpdateTenantRequest defines the request body for editing a tenant
type UpdateTenantRequest struct {
	Name        string `json:"name" binding:"required"`
	LevelID     string `json:"levelId"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Website     string `json:"website"`
}

// ListTenantsRequest defines query parameters for listing tenants
type ListTenantsRequest struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"pageSize,default=20"`
	Name     string `form:"name"` // Corresponds to ?name=... in Query Param
}

// ResetTenantAdminPasswordRequest defines the request body for resetting the tenant super admin's password
type ResetTenantAdminPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
