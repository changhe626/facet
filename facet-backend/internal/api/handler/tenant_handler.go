package handler

import (
	"errors"
	"facet-backend/internal/api/middleware"
	"facet-backend/internal/api/request"
	"facet-backend/internal/api/response"
	"facet-backend/internal/api/validator"
	"facet-backend/internal/domain"
	"facet-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TenantHandler struct {
	tenantService service.TenantService
}

func NewTenantHandler(tenantService service.TenantService) *TenantHandler {
	return &TenantHandler{tenantService: tenantService}
}

func (h *TenantHandler) Create(c *gin.Context) {
	var req request.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleBindingError(c, err)
		return
	}

	userRoleVal, exists := c.Get(string(middleware.UserRoleKey))
	if !exists {
		_ = c.Error(errors.New("userRole not found in context"))
		return
	}
	userIDVal, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		_ = c.Error(errors.New("userID not found in context"))
		return
	}

	roleStr, _ := userRoleVal.(string)
	if domain.UserRole(roleStr) != domain.RoleSystemAdmin {
		response.CustomError(c, http.StatusForbidden, "permission denied: only system admin can create tenants")
		return
	}

	creatorID, _ := userIDVal.(string)

	if err := h.tenantService.CreateTenantWithAdmin(c.Request.Context(), creatorID, req); err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, true)
}

func (h *TenantHandler) Update(c *gin.Context) {
	tenantID := c.Param("tenantId")

	var req request.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleBindingError(c, err)
		return
	}

	userRoleVal, exists := c.Get(string(middleware.UserRoleKey))
	if !exists {
		_ = c.Error(errors.New("userRole not found in context"))
		return
	}
	userIDVal, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		_ = c.Error(errors.New("userID not found in context"))
		return
	}

	roleStr, _ := userRoleVal.(string)
	if domain.UserRole(roleStr) != domain.RoleSystemAdmin {
		response.CustomError(c, http.StatusForbidden, "permission denied: only system admin can update tenants")
		return
	}

	updaterID, _ := userIDVal.(string)

	if err := h.tenantService.Update(c.Request.Context(), tenantID, updaterID, req); err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, true)
}

func (h *TenantHandler) List(c *gin.Context) {
	var req request.ListTenantsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.CustomError(c, http.StatusBadRequest, "invalid query parameters")
		return
	}

	userRoleVal, exists := c.Get(string(middleware.UserRoleKey))
	if !exists {
		_ = c.Error(errors.New("userRole not found in context"))
		return
	}
	roleStr, _ := userRoleVal.(string)
	if domain.UserRole(roleStr) != domain.RoleSystemAdmin {
		response.CustomError(c, http.StatusForbidden, "permission denied: only system admin can list tenants")
		return
	}

	resp, err := h.tenantService.List(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, resp)
}

func (h *TenantHandler) ResetAdminPassword(c *gin.Context) {
	tenantID := c.Param("tenantId")

	var req request.ResetTenantAdminPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleBindingError(c, err)
		return
	}

	userRoleVal, exists := c.Get(string(middleware.UserRoleKey))
	if !exists {
		_ = c.Error(errors.New("userRole not found in context"))
		return
	}
	roleStr, _ := userRoleVal.(string)
	if domain.UserRole(roleStr) != domain.RoleSystemAdmin {
		response.CustomError(c, http.StatusForbidden, "permission denied: only system admin can reset tenant admin password")
		return
	}

	if err := h.tenantService.ResetTenantAdminPassword(c.Request.Context(), tenantID, req.NewPassword); err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, true)
}

func (h *TenantHandler) GetByID(c *gin.Context) {
	tenantID := c.Param("tenantId")

	userRoleVal, exists := c.Get(string(middleware.UserRoleKey))
	if !exists {
		_ = c.Error(errors.New("userRole not found in context"))
		return
	}
	roleStr, _ := userRoleVal.(string)
	if domain.UserRole(roleStr) != domain.RoleSystemAdmin {
		response.CustomError(c, http.StatusForbidden, "permission denied: only system admin can get tenant details")
		return
	}

	tenant, err := h.tenantService.GetByID(c.Request.Context(), tenantID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, tenant)
}
