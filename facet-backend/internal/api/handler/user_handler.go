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

const defaultPassword = "12345678"

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) List(c *gin.Context) {
	var req request.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.CustomError(c, http.StatusBadRequest, "invalid query parameters")
		return
	}
	tenantIDVal, exists := c.Get(string(middleware.TenantIDKey))
	if !exists {
		_ = c.Error(errors.New("tenantID not found in context"))
		return
	}
	userIDVal, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		_ = c.Error(errors.New("userID not found in context"))
		return
	}
	userRoleVal, exists := c.Get(string(middleware.UserRoleKey))
	if !exists {
		_ = c.Error(errors.New("userRole not found in context"))
		return
	}
	tenantID, ok := tenantIDVal.(string)
	if !ok {
		_ = c.Error(errors.New("tenantID in context is not a string"))
		return
	}
	userID, ok := userIDVal.(string)
	if !ok {
		_ = c.Error(errors.New("userID in context is not a string"))
		return
	}
	roleStr, ok := userRoleVal.(string)
	if !ok {
		_ = c.Error(errors.New("userRole in context is not a string"))
		return
	}
	role := domain.UserRole(roleStr)
	resp, err := h.userService.ListUsers(c.Request.Context(), tenantID, userID, role, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Success(c, resp)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req request.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleBindingError(c, err)
		return
	}
	tenantIDVal, exists := c.Get(string(middleware.TenantIDKey))
	if !exists {
		_ = c.Error(errors.New("tenantID not found in context"))
		return
	}
	creatorIDVal, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		_ = c.Error(errors.New("creatorID not found in context"))
		return
	}
	userRoleVal, exists := c.Get(string(middleware.UserRoleKey))
	if !exists {
		_ = c.Error(errors.New("userRole not found in context"))
		return
	}

	roleStr, _ := userRoleVal.(string)
	creatorRole := domain.UserRole(roleStr)
	switch creatorRole {
	case domain.RoleSystemAdmin, domain.RoleTenantSuperAdmin, domain.RoleTenantAdmin:
	default:
		response.CustomError(c, http.StatusForbidden, "permission denied: insufficient privileges to create users")
		return
	}

	tenantID, _ := tenantIDVal.(string)
	creatorID, _ := creatorIDVal.(string)

	resp, err := h.userService.Create(c.Request.Context(), tenantID, creatorID, req)
	if err != nil {
		if errors.Is(err, service.ErrQuotaExceededTenantAdmins) || errors.Is(err, service.ErrQuotaExceededTenantViewers) {
			response.CustomError(c, http.StatusForbidden, err.Error())
			return
		}
		if errors.Is(err, domain.ErrEmailExistsInTenant) {
			response.CustomError(c, http.StatusConflict, err.Error())
			return
		}
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, response.ApiResponse{
		Code:    http.StatusCreated,
		Message: "Created",
		Data:    resp,
	})
}

func (h *UserHandler) Delete(c *gin.Context) {
	userIDToDelete := c.Param("userId")

	userRoleVal, exists := c.Get(string(middleware.UserRoleKey))
	if !exists {
		_ = c.Error(errors.New("userRole not found in context"))
		return
	}

	roleStr, _ := userRoleVal.(string)
	deleterRole := domain.UserRole(roleStr)

	switch deleterRole {
	case domain.RoleSystemAdmin, domain.RoleTenantSuperAdmin, domain.RoleTenantAdmin:
	default:
		response.CustomError(c, http.StatusForbidden, "permission denied: insufficient privileges to delete users")
		return
	}

	err := h.userService.Delete(c.Request.Context(), userIDToDelete)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, true)
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	userIDToReset := c.Param("userId")
	userRoleVal, exists := c.Get(string(middleware.UserRoleKey))
	if !exists {
		_ = c.Error(errors.New("userRole not found in context"))
		return
	}

	roleStr, _ := userRoleVal.(string)
	creatorRole := domain.UserRole(roleStr)
	switch creatorRole {
	case domain.RoleSystemAdmin, domain.RoleTenantSuperAdmin, domain.RoleTenantAdmin:
	default:
		response.CustomError(c, http.StatusForbidden, "permission denied: insufficient privileges to reset password")
		return
	}

	err := h.userService.ResetPassword(c.Request.Context(), userIDToReset, defaultPassword)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Success(c, true)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	userIDToGet := c.Param("userId")
	tenantIDVal, exists := c.Get(string(middleware.TenantIDKey))
	if !exists {
		_ = c.Error(errors.New("tenantID not found in context"))
		return
	}
	tenantID, ok := tenantIDVal.(string)
	if !ok {
		_ = c.Error(errors.New("tenantID in context is not a string"))
		return
	}
	resp, err := h.userService.GetByID(c.Request.Context(), tenantID, userIDToGet)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, resp)
}

func (h *UserHandler) Update(c *gin.Context) {
	tenantIDVal, exists := c.Get(string(middleware.TenantIDKey))
	if !exists {
		_ = c.Error(errors.New("tenantID not found in context"))
		return
	}
	updaterIDVal, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		_ = c.Error(errors.New("updaterID not found in context"))
		return
	}
	tenantID := tenantIDVal.(string)
	updaterID := updaterIDVal.(string)

	userID := c.Param("userId")

	var req request.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleBindingError(c, err)
		return
	}

	if err := h.userService.Update(c.Request.Context(), tenantID, userID, updaterID, req); err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, true)
}
