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

type EnvironmentHandler struct {
	envService service.EnvironmentService
}

func NewEnvironmentHandler(envService service.EnvironmentService) *EnvironmentHandler {
	return &EnvironmentHandler{envService: envService}
}

func (h *EnvironmentHandler) ListActive(c *gin.Context) {
	tenantID, exists := c.Get(string(middleware.TenantIDKey))
	if !exists {
		_ = c.Error(errors.New("tenantID not found in context"))
		return
	}

	resp, err := h.envService.ListActive(c.Request.Context(), tenantID.(string))
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, resp)
}

func (h *EnvironmentHandler) List(c *gin.Context) {
	tenantID, exists := c.Get(string(middleware.TenantIDKey))
	if !exists {
		_ = c.Error(errors.New("tenantID not found in context"))
		return
	}

	var req request.ListEnvironmentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.CustomError(c, http.StatusBadRequest, "invalid query parameters")
		return
	}

	resp, err := h.envService.List(c.Request.Context(), tenantID.(string), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, resp)
}

func (h *EnvironmentHandler) Create(c *gin.Context) {
	tenantID, exists := c.Get(string(middleware.TenantIDKey))
	if !exists {
		_ = c.Error(errors.New("tenantID not found in context"))
		return
	}
	creatorID, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		_ = c.Error(errors.New("creatorID not found in context"))
		return
	}

	var req request.CreateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleBindingError(c, err)
		return
	}

	resp, err := h.envService.Create(c.Request.Context(), tenantID.(string), creatorID.(string), req)
	if err != nil {
		if errors.Is(err, service.ErrQuotaExceededEnvironments) || errors.Is(err, domain.ErrEnvironmentNameExists) {
			response.CustomError(c, http.StatusForbidden, err.Error())
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

func (h *EnvironmentHandler) Update(c *gin.Context) {
	tenantID, exists := c.Get(string(middleware.TenantIDKey))
	if !exists {
		_ = c.Error(errors.New("tenantID not found in context"))
		return
	}
	updaterID, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		_ = c.Error(errors.New("updaterID not found in context"))
		return
	}
	envID := c.Param("envId")

	var req request.UpdateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleBindingError(c, err)
		return
	}

	resp, err := h.envService.Update(c.Request.Context(), tenantID.(string), envID, updaterID.(string), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, resp)
}

func (h *EnvironmentHandler) Delete(c *gin.Context) {
	tenantID, exists := c.Get(string(middleware.TenantIDKey))
	if !exists {
		_ = c.Error(errors.New("tenantID not found in context"))
		return
	}
	envID := c.Param("envId")

	if err := h.envService.Delete(c.Request.Context(), tenantID.(string), envID); err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, true)
}
