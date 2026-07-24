package domain

import (
	"context"
	"database/sql"
	"time"
)

// NotificationLevel defines the level of a notification
type NotificationLevel string

const (
	LevelNormal    NotificationLevel = "normal"
	LevelImportant NotificationLevel = "important"
	LevelUrgent    NotificationLevel = "urgent"
)

// NotificationType defines the type of a notification
type NotificationType string

const (
	TypeAnnouncement  NotificationType = "announcement"
	TypeProductUpdate NotificationType = "product_update"
	TypeSecurityAlert NotificationType = "security_alert"
)

// SystemNotification corresponds to the system_notifications table, storing notification template content
type SystemNotification struct {
	ID        string            `db:"id"`
	Level     NotificationLevel `db:"level"`
	Type      NotificationType  `db:"type"`
	Title     string            `db:"title"`
	Content   string            `db:"content"`
	DeletedAt sql.NullTime      `db:"deleted_at"`
	CreatedAt time.Time         `db:"created_at"`
}

// UserNotification corresponds to the user_notifications table, storing user read status
type UserNotification struct {
	ID             string       `db:"id"`
	TenantID       string       `db:"tenant_id"`
	UserID         string       `db:"user_id"`
	NotificationID string       `db:"notification_id"`
	IsRead         bool         `db:"is_read"`
	ReadAt         sql.NullTime `db:"read_at"`
	CreatedAt      time.Time    `db:"created_at"`
}

// ListSystemNotificationsParams defines the parameters for querying system notifications
type ListSystemNotificationsParams struct {
	Limit  int
	Offset int
}

// NotificationRepository defines the interface for notification data operations
type NotificationRepository interface {
	// --- User side ---
	// ListAllSystem gets all system notifications that have not been soft-deleted
	ListAllSystem(ctx context.Context) ([]*SystemNotification, error)
	// FindUserNotifications gets all existing notification status records for a user
	FindUserNotifications(ctx context.Context, userID string) ([]*UserNotification, error)
	// MarkAsRead marks a batch of notifications as read for a user (creates the record if it does not exist)
	MarkAsRead(ctx context.Context, tenantID, userID string, notificationIDs []string) error
	// CountAllSystem counts all valid system notifications
	CountAllSystem(ctx context.Context) (int64, error)
	// CountReadForUser counts the total number of valid notifications read by a user
	CountReadForUser(ctx context.Context, userID string) (int64, error)

	// --- Admin side ---
	CreateSystemNotification(ctx context.Context, notification *SystemNotification) error
	UpdateSystemNotification(ctx context.Context, notification *SystemNotification) error
	DeleteSystemNotification(ctx context.Context, notificationID string) error
	FindSystemNotificationByID(ctx context.Context, notificationID string) (*SystemNotification, error)
	ListAndCountSystemNotifications(ctx context.Context, params ListSystemNotificationsParams) ([]*SystemNotification, int64, error)
}
