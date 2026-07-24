package handler

import (
	"facet-backend/internal/api/request"
	"facet-backend/internal/api/response"
	"facet-backend/internal/api/validator"
	"facet-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SystemNotificationHandler struct {
	notifService service.NotificationService
}

func NewSystemNotificationHandler(notifService service.NotificationService) *SystemNotificationHandler {
	return &SystemNotificationHandler{notifService: notifService}
}

func (h *SystemNotificationHandler) Create(c *gin.Context) {
	var req request.CreateSystemNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleBindingError(c, err)
		return
	}

	notifReq := service.SystemNotificationResponse{
		Level:   req.Level,
		Type:    req.Type,
		Title:   req.Title,
		Content: req.Content,
	}

	resp, err := h.notifService.CreateSystemNotification(c.Request.Context(), notifReq)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, resp)
}

func (h *SystemNotificationHandler) Update(c *gin.Context) {
	notificationID := c.Param("notificationId")
	var req request.UpdateSystemNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleBindingError(c, err)
		return
	}

	notifReq := service.SystemNotificationResponse{
		Level:   req.Level,
		Type:    req.Type,
		Title:   req.Title,
		Content: req.Content,
	}

	resp, err := h.notifService.UpdateSystemNotification(c.Request.Context(), notificationID, notifReq)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, resp)
}

func (h *SystemNotificationHandler) Delete(c *gin.Context) {
	notificationID := c.Param("notificationId")
	if err := h.notifService.DeleteSystemNotification(c.Request.Context(), notificationID); err != nil {
		_ = c.Error(err)
		return
	}
	response.Success(c, true)
}

func (h *SystemNotificationHandler) GetByID(c *gin.Context) {
	notificationID := c.Param("notificationId")
	resp, err := h.notifService.FindSystemNotificationByID(c.Request.Context(), notificationID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	response.Success(c, resp)
}

func (h *SystemNotificationHandler) List(c *gin.Context) {
	var req request.ListSystemNotificationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.CustomError(c, http.StatusBadRequest, "invalid query parameters")
		return
	}

	resp, err := h.notifService.ListSystemNotifications(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, resp)
}
