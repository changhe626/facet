package service

import (
	"context"
	"facet-backend/internal/domain"
	"math"
	"time"

	"github.com/bwmarrin/snowflake"
)

type NotificationStatusResponse struct {
	TotalCount  int64 `json:"totalCount"`
	UnreadCount int64 `json:"unreadCount"`
}

type UserNotificationView struct {
	ID        string                   `json:"id"`
	Level     domain.NotificationLevel `json:"level"`
	Type      domain.NotificationType  `json:"type"`
	Title     string                   `json:"title"`
	Content   string                   `json:"content"`
	CreatedAt time.Time                `json:"createdAt"`
	IsRead    bool                     `json:"isRead"`
	ReadAt    *time.Time               `json:"readAt,omitempty"`
}

type SystemNotificationResponse struct {
	ID        string    `json:"id"`
	Level     string    `json:"level"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type ListSystemNotificationsResponse struct {
	Page       int                          `json:"page"`
	PageSize   int                          `json:"pageSize"`
	Total      int64                        `json:"total"`
	TotalPages int                          `json:"totalPages"`
	Items      []SystemNotificationResponse `json:"items"`
}

type NotificationService interface {
	ListForUser(ctx context.Context, userID string) ([]*UserNotificationView, error)
	MarkAsRead(ctx context.Context, tenantID, userID string, notificationIDs []string) error
	GetStatus(ctx context.Context, userID string) (*NotificationStatusResponse, error)

	CreateSystemNotification(ctx context.Context, req SystemNotificationResponse) (*SystemNotificationResponse, error)
	UpdateSystemNotification(ctx context.Context, id string, req SystemNotificationResponse) (*SystemNotificationResponse, error)
	DeleteSystemNotification(ctx context.Context, id string) error
	FindSystemNotificationByID(ctx context.Context, id string) (*SystemNotificationResponse, error)
	ListSystemNotifications(ctx context.Context, page, pageSize int) (*ListSystemNotificationsResponse, error)
}

type notificationService struct {
	notifRepo     domain.NotificationRepository
	snowflakeNode *snowflake.Node
}

func NewNotificationService(notifRepo domain.NotificationRepository, node *snowflake.Node) NotificationService {
	return &notificationService{
		notifRepo:     notifRepo,
		snowflakeNode: node,
	}
}

func (s *notificationService) ListForUser(ctx context.Context, userID string) ([]*UserNotificationView, error) {
	systemNotifs, err := s.notifRepo.ListAllSystem(ctx)
	if err != nil {
		return nil, err
	}

	userNotifs, err := s.notifRepo.FindUserNotifications(ctx, userID)
	if err != nil {
		return nil, err
	}

	readMap := make(map[string]*domain.UserNotification, len(userNotifs))
	for _, un := range userNotifs {
		readMap[un.NotificationID] = un
	}

	views := make([]*UserNotificationView, 0, len(systemNotifs))
	for _, sn := range systemNotifs {
		un, exists := readMap[sn.ID]
		isRead := exists && un.IsRead
		var readAt *time.Time
		if exists && un.ReadAt.Valid {
			readAt = &un.ReadAt.Time
		}
		views = append(views, &UserNotificationView{
			ID:        sn.ID,
			Level:     sn.Level,
			Type:      sn.Type,
			Title:     sn.Title,
			Content:   sn.Content,
			CreatedAt: sn.CreatedAt,
			IsRead:    isRead,
			ReadAt:    readAt,
		})
	}
	return views, nil
}

func (s *notificationService) MarkAsRead(ctx context.Context, tenantID, userID string, notificationIDs []string) error {
	return s.notifRepo.MarkAsRead(ctx, tenantID, userID, notificationIDs)
}

func (s *notificationService) GetStatus(ctx context.Context, userID string) (*NotificationStatusResponse, error) {
	totalCount, err := s.notifRepo.CountAllSystem(ctx)
	if err != nil {
		return nil, err
	}

	readCount, err := s.notifRepo.CountReadForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	unreadCount := totalCount - readCount
	if unreadCount < 0 {
		unreadCount = 0
	}

	return &NotificationStatusResponse{
		TotalCount:  totalCount,
		UnreadCount: unreadCount,
	}, nil
}

func (s *notificationService) CreateSystemNotification(ctx context.Context, req SystemNotificationResponse) (*SystemNotificationResponse, error) {
	notif := &domain.SystemNotification{
		ID:      "notif-" + s.snowflakeNode.Generate().String(),
		Level:   domain.NotificationLevel(req.Level),
		Type:    domain.NotificationType(req.Type),
		Title:   req.Title,
		Content: req.Content,
	}
	if err := s.notifRepo.CreateSystemNotification(ctx, notif); err != nil {
		return nil, err
	}
	return s.FindSystemNotificationByID(ctx, notif.ID)
}

func (s *notificationService) UpdateSystemNotification(ctx context.Context, id string, req SystemNotificationResponse) (*SystemNotificationResponse, error) {
	notif := &domain.SystemNotification{
		ID:      id,
		Level:   domain.NotificationLevel(req.Level),
		Type:    domain.NotificationType(req.Type),
		Title:   req.Title,
		Content: req.Content,
	}
	if err := s.notifRepo.UpdateSystemNotification(ctx, notif); err != nil {
		return nil, err
	}
	return s.FindSystemNotificationByID(ctx, id)
}

func (s *notificationService) DeleteSystemNotification(ctx context.Context, id string) error {
	return s.notifRepo.DeleteSystemNotification(ctx, id)
}

func (s *notificationService) FindSystemNotificationByID(ctx context.Context, id string) (*SystemNotificationResponse, error) {
	notif, err := s.notifRepo.FindSystemNotificationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &SystemNotificationResponse{
		ID:        notif.ID,
		Level:     string(notif.Level),
		Type:      string(notif.Type),
		Title:     notif.Title,
		Content:   notif.Content,
		CreatedAt: notif.CreatedAt,
	}, nil
}

func (s *notificationService) ListSystemNotifications(ctx context.Context, page, pageSize int) (*ListSystemNotificationsResponse, error) {
	params := domain.ListSystemNotificationsParams{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}
	notifs, total, err := s.notifRepo.ListAndCountSystemNotifications(ctx, params)
	if err != nil {
		return nil, err
	}

	items := make([]SystemNotificationResponse, len(notifs))
	for i, notif := range notifs {
		items[i] = SystemNotificationResponse{
			ID:        notif.ID,
			Level:     string(notif.Level),
			Type:      string(notif.Type),
			Title:     notif.Title,
			Content:   notif.Content,
			CreatedAt: notif.CreatedAt,
		}
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(pageSize)))
	}

	return &ListSystemNotificationsResponse{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		Items:      items,
	}, nil
}
