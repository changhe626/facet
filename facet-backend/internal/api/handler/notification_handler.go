package handler

import (
	"errors"
	"facet-backend/internal/api/middleware"
	"facet-backend/internal/api/request"
	"facet-backend/internal/api/response"
	"facet-backend/internal/api/validator"
	"facet-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notifService service.NotificationService
}

func NewNotificationHandler(notifService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifService: notifService}
}

func (h *NotificationHandler) ListForUser(c *gin.Context) {
	userID, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		_ = c.Error(errors.New("userID not found in context"))
		return
	}

	notifications, err := h.notifService.ListForUser(c.Request.Context(), userID.(string))
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, notifications)
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, userExists := c.Get(string(middleware.UserIDKey))
	tenantID, tenantExists := c.Get(string(middleware.TenantIDKey))
	if !userExists || !tenantExists {
		_ = c.Error(errors.New("user or tenant information not found in context"))
		return
	}

	var req request.MarkAsReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleBindingError(c, err)
		return
	}

	if err := h.notifService.MarkAsRead(c.Request.Context(), tenantID.(string), userID.(string), req.NotificationIDs); err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, true)
}

func (h *NotificationHandler) GetStatus(c *gin.Context) {
	userID, exists := c.Get(string(middleware.UserIDKey))
	if !exists {
		_ = c.Error(errors.New("userID not found in context"))
		return
	}

	status, err := h.notifService.GetStatus(c.Request.Context(), userID.(string))
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, status)
}
