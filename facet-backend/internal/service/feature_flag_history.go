package service

import (
	"context"
	"encoding/json"
	"errors"
	"facet-backend/internal/api/response"
	"facet-backend/internal/domain"
	"fmt"
	"log/slog"
	"time"
)

func (s *featureFlagService) createHistoryEntry(ctx context.Context, flagID, envID, userID string, changeType domain.ChangeType) error {
	flag, err := s.ffRepo.GetByID(ctx, flagID, envID)
	if err != nil {
		if !errors.Is(err, domain.ErrFeatureFlagNotFound) {
			return fmt.Errorf("failed to create history entry: unable to get flag entity: %w", err)
		}
		flag = &domain.FeatureFlag{ID: flagID, Name: "(deleted)"}
	}

	rules, err := s.ruleRepo.ListByFeatureFlagID(ctx, flagID)
	if err != nil {
		return fmt.Errorf("failed to create history entry: unable to get rules: %w", err)
	}

	simpleRules := make([]response.RuleSnapshot, len(rules))
	for i, r := range rules {
		simpleRules[i] = response.RuleSnapshot{
			AccountID: r.AccountID.String,
			UserIDs:   r.UserIDs,
		}
	}
	rulesSnapshot, err := json.Marshal(simpleRules)
	if err != nil {
		return fmt.Errorf("failed to create history entry: failed to serialize rule snapshot: %w", err)
	}

	history := &domain.FeatureFlagHistory{
		HistoryID:         "ffh-" + s.snowflakeNode.Generate().String(),
		FeatureFlagID:     flag.ID,
		ChangeTime:        time.Now(),
		ChangeType:        changeType,
		ChangedBy:         userID,
		Name:              flag.Name,
		Description:       flag.Description.String,
		Enabled:           flag.Enabled,
		FlagType:          string(flag.FlagType),
		RolloutPercentage: flag.RolloutPercentage,
		ExpiresAt:         flag.ExpiresAt,
		RulesSnapshot:     rulesSnapshot,
	}

	return s.historyRepo.Create(ctx, history)
}

func (s *featureFlagService) GetHistory(ctx context.Context, envID, flagID string) ([]*response.FeatureFlagSnapshotResponse, error) {
	histories, err := s.historyRepo.ListByFeatureFlagID(ctx, flagID)
	if err != nil {
		return nil, fmt.Errorf("failed to get history records: %w", err)
	}

	if len(histories) == 0 {
		return []*response.FeatureFlagSnapshotResponse{}, nil
	}

	result := make([]*response.FeatureFlagSnapshotResponse, len(histories))

	for i, current := range histories {
		var rules []response.RuleSnapshot
		if err := json.Unmarshal(current.RulesSnapshot, &rules); err != nil {
			slog.Warn("failed to parse rule snapshot", "historyID", current.HistoryID, "error", err)
			rules = []response.RuleSnapshot{}
		}

		snapshot := &response.FeatureFlagSnapshotResponse{
			Name:           current.Name,
			Description:    current.Description,
			Enabled:        current.Enabled,
			FlagType:       current.FlagType,
			ExpiresAt:      current.ExpiresAt.String,
			WhitelistRules: rules,
			HistoryID:      current.HistoryID,
			ChangeTime:     current.ChangeTime,
			ChangeType:     current.ChangeType,
			ChangedBy:      current.ChangedBy,
		}
		result[i] = snapshot
	}

	return result, nil
}
