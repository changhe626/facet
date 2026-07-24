package service

import (
	"context"
	"database/sql"
	"fmt"

	"facet-backend/internal/api/request"
	"facet-backend/internal/domain"

	"github.com/bwmarrin/snowflake"
)

type SegmentService interface {
	List(ctx context.Context, tenantID string, keyword string, page, pageSize int) ([]*domain.Segment, int64, error)
	Create(ctx context.Context, tenantID, userID string, req request.CreateSegmentRequest) (*domain.Segment, error)
	GetByID(ctx context.Context, id, tenantID string) (*domain.Segment, error)
	Update(ctx context.Context, tenantID, userID string, req request.UpdateSegmentRequest) (*domain.Segment, error)
	Delete(ctx context.Context, id, tenantID string) error
}

type segmentService struct {
	segRepo     domain.SegmentRepository
	condRepo    domain.SegmentConditionRepository
	sf          *snowflake.Node
}

func NewSegmentService(segRepo domain.SegmentRepository, condRepo domain.SegmentConditionRepository, sf *snowflake.Node) SegmentService {
	return &segmentService{segRepo: segRepo, condRepo: condRepo, sf: sf}
}

func (s *segmentService) List(ctx context.Context, tenantID string, keyword string, page, pageSize int) ([]*domain.Segment, int64, error) {
	offset := (page - 1) * pageSize
	segments, total, err := s.segRepo.ListAndCount(ctx, domain.ListSegmentsParams{
		TenantID: tenantID,
		Keyword:  keyword,
		Limit:    pageSize,
		Offset:   offset,
	})
	if err != nil {
		return nil, 0, err
	}
	for _, seg := range segments {
		conds, err := s.condRepo.ListBySegmentID(ctx, seg.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to load conditions for segment %s: %w", seg.ID, err)
		}
		for _, c := range conds {
			seg.Conditions = append(seg.Conditions, *c)
		}
	}
	return segments, total, nil
}

func (s *segmentService) Create(ctx context.Context, tenantID, userID string, req request.CreateSegmentRequest) (*domain.Segment, error) {
	existingID, err := s.segRepo.CheckNameExists(ctx, req.Name, tenantID)
	if err != nil {
		return nil, err
	}
	if existingID != "" {
		return nil, domain.ErrSegmentNameExists
	}

	matchType := domain.SegmentMatchAny
	if req.MatchType == "ALL" {
		matchType = domain.SegmentMatchAll
	}

	seg := &domain.Segment{
		ID:          "seg-" + s.sf.Generate().String(),
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		MatchType:   matchType,
		CreatedBy:   sql.NullString{String: userID, Valid: true},
		UpdatedBy:   sql.NullString{String: userID, Valid: true},
	}

	if err := s.segRepo.Create(ctx, seg); err != nil {
		return nil, err
	}

	for _, c := range req.Conditions {
		cond := &domain.SegmentCondition{
			ID:           "segc-" + s.sf.Generate().String(),
			SegmentID:    seg.ID,
			AttributeKey: c.AttributeKey,
			Operator:     c.Operator,
			Value:        c.Value,
		}
		if err := s.condRepo.Create(ctx, cond); err != nil {
			return nil, err
		}
		seg.Conditions = append(seg.Conditions, *cond)
	}

	return seg, nil
}

func (s *segmentService) GetByID(ctx context.Context, id, tenantID string) (*domain.Segment, error) {
	seg, err := s.segRepo.GetByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}
	conds, err := s.condRepo.ListBySegmentID(ctx, seg.ID)
	if err != nil {
		return nil, err
	}
	for _, c := range conds {
		seg.Conditions = append(seg.Conditions, *c)
	}
	return seg, nil
}

func (s *segmentService) Update(ctx context.Context, tenantID, userID string, req request.UpdateSegmentRequest) (*domain.Segment, error) {
	existingID, err := s.segRepo.CheckNameExists(ctx, req.Name, tenantID)
	if err != nil {
		return nil, err
	}
	if existingID != "" && existingID != req.ID {
		return nil, domain.ErrSegmentNameExists
	}

	matchType := domain.SegmentMatchAny
	if req.MatchType == "ALL" {
		matchType = domain.SegmentMatchAll
	}

	seg := &domain.Segment{
		ID:          req.ID,
		TenantID:    tenantID,
		Name:        req.Name,
		Description: req.Description,
		MatchType:   matchType,
		UpdatedBy:   sql.NullString{String: userID, Valid: true},
	}

	if err := s.segRepo.Update(ctx, seg); err != nil {
		return nil, err
	}

	if err := s.condRepo.DeleteBySegmentID(ctx, req.ID); err != nil {
		return nil, err
	}

	for _, c := range req.Conditions {
		cond := &domain.SegmentCondition{
			ID:           "segc-" + s.sf.Generate().String(),
			SegmentID:    req.ID,
			AttributeKey: c.AttributeKey,
			Operator:     c.Operator,
			Value:        c.Value,
		}
		if err := s.condRepo.Create(ctx, cond); err != nil {
			return nil, err
		}
		seg.Conditions = append(seg.Conditions, *cond)
	}

	return seg, nil
}

func (s *segmentService) Delete(ctx context.Context, id, tenantID string) error {
	if err := s.segRepo.Delete(ctx, id, tenantID); err != nil {
		return err
	}
	return s.condRepo.DeleteBySegmentID(ctx, id)
}
