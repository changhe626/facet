package domain

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrSegmentNotFound = errors.New("segment not found")
	ErrSegmentNameExists = errors.New("segment name already exists in this tenant")
)

type SegmentMatchType string

const (
	SegmentMatchAny SegmentMatchType = "ANY"
	SegmentMatchAll SegmentMatchType = "ALL"
)

type Segment struct {
	ID          string            `db:"id"`
	TenantID    string            `db:"tenant_id"`
	Name        string            `db:"name"`
	Description string            `db:"description"`
	MatchType   SegmentMatchType  `db:"match_type"`
	CreatedBy   sql.NullString    `db:"created_by"`
	UpdatedBy   sql.NullString    `db:"updated_by"`
	CreatedAt   time.Time         `db:"created_at"`
	UpdatedAt   time.Time         `db:"updated_at"`
	DeletedAt   sql.NullTime      `db:"deleted_at"`
	Conditions  []SegmentCondition
}

type SegmentCondition struct {
	ID           string `db:"id"`
	SegmentID    string `db:"segment_id"`
	AttributeKey string `db:"attribute_key"`
	Operator     string `db:"operator"`
	Value        string `db:"value"`
}

type ListSegmentsParams struct {
	TenantID string
	Keyword  string
	Limit    int
	Offset   int
}

type SegmentRepository interface {
	ListAndCount(ctx context.Context, params ListSegmentsParams) ([]*Segment, int64, error)
	Create(ctx context.Context, segment *Segment) error
	GetByID(ctx context.Context, id string, tenantID string) (*Segment, error)
	Update(ctx context.Context, segment *Segment) error
	Delete(ctx context.Context, id string, tenantID string) error
	CheckNameExists(ctx context.Context, name, tenantID string) (string, error)
}

type SegmentConditionRepository interface {
	Create(ctx context.Context, cond *SegmentCondition) error
	ListBySegmentID(ctx context.Context, segmentID string) ([]*SegmentCondition, error)
	DeleteBySegmentID(ctx context.Context, segmentID string) error
}

type FeatureFlagSegmentRepository interface {
	SetSegments(ctx context.Context, flagID string, segmentIDs []string) error
	GetSegmentIDs(ctx context.Context, flagID string) ([]string, error)
}
