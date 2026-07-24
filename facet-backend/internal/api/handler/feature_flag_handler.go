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

type FeatureFlagHandler struct {
	ffService service.FeatureFlagService
}

func NewFeatureFlagHandler(ffService service.FeatureFlagService) *FeatureFlagHandler {
	return &FeatureFlagHandler{ffService: ffService}
}

func (h *FeatureFlagHandler) List(c *gin.Context) {
	envID := c.Param("envId")
	var req request.ListFeatureFlagsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.CustomError(c, http.StatusBadRequest, "invalid query parameters")
		return
	}
	resp, err := h.ffService.List(c.Request.Context(), envID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Success(c, resp)
}

func (h *FeatureFlagHandler) Create(c *gin.Context) {
	envID := c.Param("envId")
	var req request.CreateFeatureFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleBindingError(c, err)
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
	tenantID, _ := tenantIDVal.(string)
	userID, _ := userIDVal.(string)
	resp, err := h.ffService.Create(c.Request.Context(), tenantID, envID, userID, req)
	if err != nil {
		if errors.Is(err, service.ErrQuotaExceededFeatureFlags) || errors.Is(err, domain.ErrFeatureFlagNameExistsInTenant) {
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

func (h *FeatureFlagHandler) GetByID(c *gin.Context) {
	envID := c.Param("envId")
	featureID := c.Param("featureId")
	resp, err := h.ffService.GetByID(c.Request.Context(), envID, featureID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Success(c, resp)
}

func (h *FeatureFlagHandler) Update(c *gin.Context) {
	envID := c.Param("envId")
	featureID := c.Param("featureId")

	var req request.UpdateFeatureFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleBindingError(c, err)
		return
	}

	userIDVal, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		_ = c.Error(errors.New("userID not found in context"))
		return
	}
	userID, _ := userIDVal.(string)

	resp, err := h.ffService.Update(c.Request.Context(), envID, userID, featureID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, resp)
}

func (h *FeatureFlagHandler) Delete(c *gin.Context) {
	envID := c.Param("envId")
	featureID := c.Param("featureId")

	userIDVal, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		_ = c.Error(errors.New("userID not found in context"))
		return
	}
	userID, _ := userIDVal.(string)

	err := h.ffService.Delete(c.Request.Context(), envID, featureID, userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, true)
}

func (h *FeatureFlagHandler) GetHistory(c *gin.Context) {
	envID := c.Param("envId")
	featureID := c.Param("featureId")

	history, err := h.ffService.GetHistory(c.Request.Context(), envID, featureID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, history)
}

func (h *FeatureFlagHandler) Archive(c *gin.Context) {
	envID := c.Param("envId")
	featureID := c.Param("featureId")
	userIDVal, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		_ = c.Error(errors.New("userID not found in context"))
		return
	}
	userID, _ := userIDVal.(string)

	if err := h.ffService.Archive(c.Request.Context(), envID, featureID, userID); err != nil {
		_ = c.Error(err)
		return
	}
	response.Success(c, gin.H{"archived": true})
}

func (h *FeatureFlagHandler) Unarchive(c *gin.Context) {
	envID := c.Param("envId")
	featureID := c.Param("featureId")
	userIDVal, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		_ = c.Error(errors.New("userID not found in context"))
		return
	}
	userID, _ := userIDVal.(string)

	if err := h.ffService.Unarchive(c.Request.Context(), envID, featureID, userID); err != nil {
		_ = c.Error(err)
		return
	}
	response.Success(c, gin.H{"archived": false})
}
