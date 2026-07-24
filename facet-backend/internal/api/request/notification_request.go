package request

// MarkAsReadRequest defines the request body for marking notifications as read
type MarkAsReadRequest struct {
	NotificationIDs []string `json:"notificationIds" binding:"required,gt=0"`
}
