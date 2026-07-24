package request

// CreateSystemNotificationRequest defines the request body for creating a system notification
type CreateSystemNotificationRequest struct {
	Level   string `json:"level" binding:"required,oneof=normal important urgent"`
	Type    string `json:"type" binding:"required,oneof=announcement product_update security_alert"`
	Title   string `json:"title" binding:"required,min=1,max=255"`
	Content string `json:"content" binding:"required"`
}

// UpdateSystemNotificationRequest defines the request body for updating a system notification
type UpdateSystemNotificationRequest struct {
	Level   string `json:"level" binding:"required,oneof=normal important urgent"`
	Type    string `json:"type" binding:"required,oneof=announcement product_update security_alert"`
	Title   string `json:"title" binding:"required,min=1,max=255"`
	Content string `json:"content" binding:"required"`
}

// ListSystemNotificationsRequest defines query parameters for listing system notifications
type ListSystemNotificationsRequest struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"pageSize,default=10"`
}
