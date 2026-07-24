package postgres

import (
	"context"
	"errors"
	"facet-backend/internal/domain"
	"fmt"

	"github.com/bwmarrin/snowflake"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// notificationRepository implements the domain.NotificationRepository interface
type notificationRepository struct {
	pool          *pgxpool.Pool
	snowflakeNode *snowflake.Node
}

// NewNotificationRepository creates a new NotificationRepository instance
func NewNotificationRepository(pool *pgxpool.Pool, node *snowflake.Node) domain.NotificationRepository {
	return &notificationRepository{
		pool:          pool,
		snowflakeNode: node,
	}
}

// --- User side methods ---

func (r *notificationRepository) ListAllSystem(ctx context.Context) ([]*domain.SystemNotification, error) {
	query := `SELECT id, level, type, title, content, created_at FROM system_notifications WHERE deleted_at IS NULL ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*domain.SystemNotification
	for rows.Next() {
		var notif domain.SystemNotification
		if err := rows.Scan(&notif.ID, &notif.Level, &notif.Type, &notif.Title, &notif.Content, &notif.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, &notif)
	}
	return notifications, rows.Err()
}

func (r *notificationRepository) FindUserNotifications(ctx context.Context, userID string) ([]*domain.UserNotification, error) {
	query := `SELECT notification_id, is_read, read_at FROM user_notifications WHERE user_id = $1`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userNotifs []*domain.UserNotification
	for rows.Next() {
		var un domain.UserNotification
		if err := rows.Scan(&un.NotificationID, &un.IsRead, &un.ReadAt); err != nil {
			return nil, err
		}
		userNotifs = append(userNotifs, &un)
	}
	return userNotifs, rows.Err()
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, tenantID, userID string, notificationIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, notifID := range notificationIDs {
		query := `
            INSERT INTO user_notifications (id, tenant_id, user_id, notification_id, is_read, read_at)
            VALUES ($1, $2, $3, $4, true, NOW())
            ON CONFLICT (user_id, notification_id) DO UPDATE
            SET is_read = true, read_at = NOW();
        `
		_, err := tx.Exec(ctx, query, "un-"+r.snowflakeNode.Generate().String(), tenantID, userID, notifID)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *notificationRepository) CountAllSystem(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM system_notifications WHERE deleted_at IS NULL`
	var count int64
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (r *notificationRepository) CountReadForUser(ctx context.Context, userID string) (int64, error) {
	query := `SELECT COUNT(*) FROM user_notifications un INNER JOIN system_notifications sn ON un.notification_id = sn.id WHERE un.user_id = $1 AND un.is_read = true AND sn.deleted_at IS NULL`
	var count int64
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

// --- Admin side methods ---

func (r *notificationRepository) CreateSystemNotification(ctx context.Context, notification *domain.SystemNotification) error {
	query := `INSERT INTO system_notifications (id, level, type, title, content) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.pool.Exec(ctx, query, notification.ID, notification.Level, notification.Type, notification.Title, notification.Content)
	return err
}

func (r *notificationRepository) UpdateSystemNotification(ctx context.Context, notification *domain.SystemNotification) error {
	query := `UPDATE system_notifications SET level = $1, type = $2, title = $3, content = $4 WHERE id = $5 AND deleted_at IS NULL`
	cmdTag, err := r.pool.Exec(ctx, query, notification.Level, notification.Type, notification.Title, notification.Content, notification.ID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("notification not found or already deleted")
	}
	return nil
}

func (r *notificationRepository) DeleteSystemNotification(ctx context.Context, notificationID string) error {
	query := `UPDATE system_notifications SET deleted_at = NOW() WHERE id = $1`
	cmdTag, err := r.pool.Exec(ctx, query, notificationID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("notification not found")
	}
	return nil
}

func (r *notificationRepository) FindSystemNotificationByID(ctx context.Context, notificationID string) (*domain.SystemNotification, error) {
	query := `SELECT id, level, type, title, content, created_at FROM system_notifications WHERE id = $1 AND deleted_at IS NULL`
	row := r.pool.QueryRow(ctx, query, notificationID)
	var notif domain.SystemNotification
	if err := row.Scan(&notif.ID, &notif.Level, &notif.Type, &notif.Title, &notif.Content, &notif.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("notification not found")
		}
		return nil, err
	}
	return &notif, nil
}

func (r *notificationRepository) ListAndCountSystemNotifications(ctx context.Context, params domain.ListSystemNotificationsParams) ([]*domain.SystemNotification, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM system_notifications WHERE deleted_at IS NULL`
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count system notifications: %w", err)
	}

	if total == 0 {
		return []*domain.SystemNotification{}, 0, nil
	}

	listQuery := `SELECT id, level, type, title, created_at FROM system_notifications WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, listQuery, params.Limit, params.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list system notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*domain.SystemNotification
	for rows.Next() {
		var notif domain.SystemNotification
		if err := rows.Scan(&notif.ID, &notif.Level, &notif.Type, &notif.Title, &notif.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan system notification row for list: %w", err)
		}
		notifications = append(notifications, &notif)
	}
	return notifications, total, rows.Err()
}
