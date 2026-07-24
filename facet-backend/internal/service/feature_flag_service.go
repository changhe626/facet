package service

import (
	"context"
	"database/sql"
	"errors"
	"facet-backend/internal/api/request"
	"facet-backend/internal/api/response"
	"facet-backend/internal/cache"
	"facet-backend/internal/domain"
	"fmt"
	"log/slog"
	"math"

	"github.com/bwmarrin/snowflake"
)

type FeatureFlagService interface {
	List(ctx context.Context, envID string, req request.ListFeatureFlagsRequest) (*response.ListFeatureFlagsResponse, error)
	Create(ctx context.Context, tenantID string, envID string, userID string, req request.CreateFeatureFlagRequest) (*response.CreateFeatureFlagResponse, error)
	GetByID(ctx context.Context, envID string, id string) (*response.FeatureFlagDetailResponse, error)
	Update(ctx context.Context, envID string, userID string, id string, req request.UpdateFeatureFlagRequest) (*response.UpdateFeatureFlagResponse, error)
	Delete(ctx context.Context, envID string, id string, userID string) error
	GetHistory(ctx context.Context, envID, flagID string) ([]*response.FeatureFlagSnapshotResponse, error)
	Archive(ctx context.Context, envID, id, userID string) error
	Unarchive(ctx context.Context, envID, id, userID string) error
}

type featureFlagService struct {
	ffRepo             domain.FeatureFlagRepository
	ruleRepo           domain.FeatureFlagRuleRepository
	historyRepo        domain.FeatureFlagHistoryRepository
	userRepo           domain.UserRepository
	envRepo            domain.EnvironmentRepository
	snowflakeNode      *snowflake.Node
	cacheService       cache.Service
	tenantLevelService TenantLevelService
	ffSegmentRepo      domain.FeatureFlagSegmentRepository
}

func NewFeatureFlagService(
	ffRepo domain.FeatureFlagRepository,
	ruleRepo domain.FeatureFlagRuleRepository,
	historyRepo domain.FeatureFlagHistoryRepository,
	userRepo domain.UserRepository,
	envRepo domain.EnvironmentRepository,
	node *snowflake.Node,
	cacheService cache.Service,
	tenantLevelService TenantLevelService,
	ffSegmentRepo domain.FeatureFlagSegmentRepository,
) FeatureFlagService {
	return &featureFlagService{
		ffRepo:             ffRepo,
		ruleRepo:           ruleRepo,
		historyRepo:        historyRepo,
		userRepo:           userRepo,
		envRepo:            envRepo,
		snowflakeNode:      node,
		cacheService:       cacheService,
		tenantLevelService: tenantLevelService,
		ffSegmentRepo:      ffSegmentRepo,
	}
}

func (s *featureFlagService) List(ctx context.Context, envID string, req request.ListFeatureFlagsRequest) (*response.ListFeatureFlagsResponse, error) {
	offset := (req.Page - 1) * req.PageSize
	params := domain.ListFeatureFlagsParams{
		EnvironmentID:   envID,
		Keyword:         req.Keyword,
		IncludeArchived: req.IncludeArchived,
		Limit:           req.PageSize,
		Offset:          offset,
	}

	flags, total, err := s.ffRepo.ListAndCountByEnvironment(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list feature flags from repository: %w", err)
	}

	if total == 0 {
		return &response.ListFeatureFlagsResponse{
			Page:       req.Page,
			PageSize:   req.PageSize,
			Total:      0,
			TotalPages: 0,
			Items:      []response.FeatureFlagListItem{},
		}, nil
	}

	userIDs := make([]string, 0)
	userIDSet := make(map[string]struct{})
	for _, flag := range flags {
		if flag.UpdatedBy.Valid && flag.UpdatedBy.String != "" {
			if _, exists := userIDSet[flag.UpdatedBy.String]; !exists {
				userIDs = append(userIDs, flag.UpdatedBy.String)
				userIDSet[flag.UpdatedBy.String] = struct{}{}
			}
		}
	}

	userMap := make(map[string]*domain.User)
	if len(userIDs) > 0 {
		users, err := s.userRepo.ListByIDs(ctx, userIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to list users by ids: %w", err)
		}
		for _, user := range users {
			userMap[user.ID] = user
		}
	}

	items := make([]response.FeatureFlagListItem, len(flags))
	for i, flag := range flags {
		item := response.FeatureFlagListItem{
			ID:                flag.ID,
			Name:              flag.Name,
			Enabled:           flag.Enabled,
			FlagType:          string(flag.FlagType),
			RolloutPercentage: flag.RolloutPercentage,
			UpdatedAt:         flag.UpdatedAt,
			ExpiresAt:         flag.ExpiresAt.String,
			UpdatedBy:         "",
		}
		if flag.ArchivedAt.Valid {
			item.ArchivedAt = &flag.ArchivedAt.Time
		}
		if flag.UpdatedBy.Valid {
			if user, ok := userMap[flag.UpdatedBy.String]; ok {
				item.UpdatedBy = user.DisplayName.String
			}
		}
		items[i] = item
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(req.PageSize)))
	}

	return &response.ListFeatureFlagsResponse{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
		Items:      items,
	}, nil
}

func (s *featureFlagService) Create(ctx context.Context, tenantID string, envID string, userID string, req request.CreateFeatureFlagRequest) (*response.CreateFeatureFlagResponse, error) {
	if err := s.tenantLevelService.CanCreateFeatureFlag(ctx, tenantID); err != nil {
		return nil, err
	}

	exists, err := s.ffRepo.CheckNameExistsInTenant(ctx, req.Name, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to check feature flag name existence: %w", err)
	}
	if exists {
		return nil, domain.ErrFeatureFlagNameExistsInTenant
	}

	newFlag := &domain.FeatureFlag{
		ID:                "ff-" + s.snowflakeNode.Generate().String(),
		TenantID:          tenantID,
		EnvironmentID:     envID,
		Name:              req.Name,
		Description:       sql.NullString{String: req.Description, Valid: req.Description != ""},
		Enabled:           req.Enabled,
		FlagType:          domain.FlagType(req.FlagType),
		RolloutPercentage: req.RolloutPercentage,
		CreatedBy:         sql.NullString{String: userID, Valid: true},
		UpdatedBy:         sql.NullString{String: userID, Valid: true},
	}

	if err := s.ffRepo.Create(ctx, newFlag); err != nil {
		return nil, fmt.Errorf("failed to create feature flag in transaction: %w", err)
	}

	for _, rule := range req.WhitelistRules {
		newRule := &domain.FeatureFlagRule{
			ID:            "ffr-" + s.snowflakeNode.Generate().String(),
			FeatureFlagID: newFlag.ID,
			AccountID:     sql.NullString{String: rule.AccountID, Valid: rule.AccountID != ""},
			UserIDs:       rule.UserIDs,
			CreatedBy:     sql.NullString{String: userID, Valid: true},
			UpdatedBy:     sql.NullString{String: userID, Valid: true},
		}
		if err := s.ruleRepo.Create(ctx, newRule); err != nil {
			return nil, fmt.Errorf("failed to create feature flag rule in transaction: %w", err)
		}
	}

	if len(req.SegmentIDs) > 0 {
		if err := s.ffSegmentRepo.SetSegments(ctx, newFlag.ID, req.SegmentIDs); err != nil {
			return nil, fmt.Errorf("failed to set segment associations: %w", err)
		}
	}

	if err := s.createHistoryEntry(ctx, newFlag.ID, envID, userID, domain.ChangeTypeCreated); err != nil {
		return nil, fmt.Errorf("failed to create initial history entry in transaction: %w", err)
	}

	userIDs := make([]string, 0)
	if newFlag.CreatedBy.Valid {
		userIDs = append(userIDs, newFlag.CreatedBy.String)
	}

	userMap := make(map[string]*domain.User)
	if len(userIDs) > 0 {
		users, err := s.userRepo.ListByIDs(ctx, userIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to list users by ids for detail view: %w", err)
		}
		for _, user := range users {
			userMap[user.ID] = user
		}
	}

	resp := &response.CreateFeatureFlagResponse{
		ID:                newFlag.ID,
		Name:              newFlag.Name,
		Description:       newFlag.Description.String,
		Enabled:           newFlag.Enabled,
		WhitelistRules:    req.WhitelistRules,
		FlagType:          string(newFlag.FlagType),
		RolloutPercentage: newFlag.RolloutPercentage,
		ExpiresAt:         newFlag.ExpiresAt.String,
		CreatedAt:         newFlag.CreatedAt,
		CreatedBy:         "",
	}
	if user, ok := userMap[newFlag.CreatedBy.String]; ok {
		resp.CreatedBy = user.DisplayName.String
	}

	s.invalidateEnvCache(ctx, envID, tenantID)

	return resp, nil
}

func (s *featureFlagService) GetByID(ctx context.Context, envID string, id string) (*response.FeatureFlagDetailResponse, error) {
	flag, err := s.ffRepo.GetByID(ctx, id, envID)
	if err != nil {
		if errors.Is(err, domain.ErrFeatureFlagNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get feature flag by id from repository: %w", err)
	}

	userIDs := make([]string, 0)
	if flag.CreatedBy.Valid {
		userIDs = append(userIDs, flag.CreatedBy.String)
	}
	if flag.UpdatedBy.Valid {
		userIDs = append(userIDs, flag.UpdatedBy.String)
	}

	userMap := make(map[string]*domain.User)
	if len(userIDs) > 0 {
		users, err := s.userRepo.ListByIDs(ctx, userIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to list users by ids for update view: %w", err)
		}
		for _, user := range users {
			userMap[user.ID] = user
		}
	}

	rules, err := s.ruleRepo.ListByFeatureFlagID(ctx, flag.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list feature flag rules: %w", err)
	}

	respRules := make([]request.WhitelistRule, len(rules))
	for i, rule := range rules {
		respRules[i] = request.WhitelistRule{
			Id:        rule.ID,
			AccountID: rule.AccountID.String,
			UserIDs:   rule.UserIDs,
		}
	}

	segIDs, _ := s.ffSegmentRepo.GetSegmentIDs(ctx, flag.ID)

	resp := &response.FeatureFlagDetailResponse{
		ID:                flag.ID,
		Name:              flag.Name,
		Description:       flag.Description.String,
		Enabled:           flag.Enabled,
		WhitelistRules:    respRules,
		FlagType:          string(flag.FlagType),
		RolloutPercentage: flag.RolloutPercentage,
		SegmentIDs:        segIDs,
		ExpiresAt:         flag.ExpiresAt.String,
		CreatedAt:         flag.CreatedAt,
		UpdatedAt:         flag.UpdatedAt,
		CreatedBy:         "",
		UpdatedBy:         "",
	}
	if user, ok := userMap[flag.CreatedBy.String]; ok {
		resp.CreatedBy = user.DisplayName.String
	}
	if user, ok := userMap[flag.UpdatedBy.String]; ok {
		resp.UpdatedBy = user.DisplayName.String
	}

	return resp, nil
}

func (s *featureFlagService) Update(ctx context.Context, envID string, userID string, id string, req request.UpdateFeatureFlagRequest) (*response.UpdateFeatureFlagResponse, error) {
	flag, err := s.ffRepo.GetByID(ctx, id, envID)
	if err != nil {
		return nil, err
	}

	if req.Description != nil {
		flag.Description = sql.NullString{String: *req.Description, Valid: true}
	}
	flag.Enabled = req.Enabled
	flag.FlagType = domain.FlagType(req.FlagType)
	flag.RolloutPercentage = req.RolloutPercentage
	flag.UpdatedBy = sql.NullString{String: userID, Valid: true}
	if err := s.ffRepo.Update(ctx, flag); err != nil {
		return nil, fmt.Errorf("failed to update feature flag in repository: %w", err)
	}

	if err := s.ruleRepo.DeleteByFeatureFlagID(ctx, id); err != nil {
		return nil, fmt.Errorf("failed to delete old feature flag rules in transaction: %w", err)
	}

	for _, rule := range req.WhitelistRules {
		newRule := &domain.FeatureFlagRule{
			ID:            "ffr-" + s.snowflakeNode.Generate().String(),
			FeatureFlagID: id,
			AccountID:     sql.NullString{String: rule.AccountID, Valid: rule.AccountID != ""},
			UserIDs:       rule.UserIDs,
			CreatedBy:     sql.NullString{String: userID, Valid: true},
			UpdatedBy:     sql.NullString{String: userID, Valid: true},
		}
		if err := s.ruleRepo.Create(ctx, newRule); err != nil {
			return nil, fmt.Errorf("failed to create new feature flag rule in transaction: %w", err)
		}
	}

	if err := s.ffSegmentRepo.SetSegments(ctx, id, req.SegmentIDs); err != nil {
		return nil, fmt.Errorf("failed to set segment associations: %w", err)
	}

	if err := s.createHistoryEntry(ctx, id, envID, userID, domain.ChangeTypeUpdated); err != nil {
		return nil, fmt.Errorf("failed to create update history entry in transaction: %w", err)
	}

	updatedFlag, err := s.ffRepo.GetByID(ctx, id, envID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated feature flag: %w", err)
	}

	resp := &response.UpdateFeatureFlagResponse{
		ID:                updatedFlag.ID,
		Name:              updatedFlag.Name,
		Description:       updatedFlag.Description.String,
		Enabled:           updatedFlag.Enabled,
		WhitelistRules:    req.WhitelistRules,
		FlagType:          string(updatedFlag.FlagType),
		RolloutPercentage: updatedFlag.RolloutPercentage,
		ExpiresAt:         updatedFlag.ExpiresAt.String,
		CreatedAt:         updatedFlag.CreatedAt,
		UpdatedAt:         updatedFlag.UpdatedAt,
	}

	s.invalidateEnvCache(ctx, envID, flag.TenantID)

	return resp, nil
}

func (s *featureFlagService) Delete(ctx context.Context, envID string, id string, userID string) error {
	flag, err := s.ffRepo.GetByID(ctx, id, envID)
	if err != nil {
		return err
	}

	tenantID := flag.TenantID

	if err := s.createHistoryEntry(ctx, flag.ID, envID, userID, domain.ChangeTypeDeleted); err != nil {
		return fmt.Errorf("failed to create delete history entry in transaction: %w", err)
	}

	if err := s.ffRepo.Delete(ctx, id, envID); err != nil {
		return fmt.Errorf("failed to delete feature flag in repository: %w", err)
	}
	if err := s.ruleRepo.DeleteByFeatureFlagID(ctx, id); err != nil {
		return fmt.Errorf("failed to delete feature flag rules in repository: %w", err)
	}

	s.invalidateEnvCache(ctx, envID, tenantID)

	return nil
}

func (s *featureFlagService) Archive(ctx context.Context, envID, id, userID string) error {
	flag, err := s.ffRepo.GetByID(ctx, id, envID)
	if err != nil {
		return err
	}
	if flag.ArchivedAt.Valid {
		return domain.ErrFeatureFlagArchived
	}

	if err := s.ffRepo.Archive(ctx, id); err != nil {
		return err
	}
	if err := s.createHistoryEntry(ctx, id, envID, userID, domain.ChangeTypeUpdated); err != nil {
		slog.Warn("failed to create history entry for archive", "flagID", id, "error", err)
	}

	s.invalidateEnvCache(ctx, envID, flag.TenantID)
	return nil
}

func (s *featureFlagService) Unarchive(ctx context.Context, envID, id, userID string) error {
	flag, err := s.ffRepo.GetByID(ctx, id, envID)
	if err != nil {
		return err
	}
	if !flag.ArchivedAt.Valid {
		return nil
	}

	if err := s.ffRepo.Unarchive(ctx, id); err != nil {
		return err
	}
	if err := s.createHistoryEntry(ctx, id, envID, userID, domain.ChangeTypeUpdated); err != nil {
		slog.Warn("failed to create history entry for unarchive", "flagID", id, "error", err)
	}

	s.invalidateEnvCache(ctx, envID, flag.TenantID)
	return nil
}
