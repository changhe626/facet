package service

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"

	"facet-backend/internal/cache"
	"facet-backend/internal/domain"
)

// EvaluationService defines the business logic interface for feature flag evaluation.
// evalContext is a JSON-encoded map of user attributes for segment targeting.
type EvaluationService interface {
	Evaluate(ctx context.Context, sdkKey, featureName, userID, accountID, appName, evalContext string) (bool, error)
}

// evaluationService implements the EvaluationService interface.
// It only depends on an abstract data source.
type evaluationService struct {
	dataSource EvaluationDataSource
}

// NewEvaluationService creates a new, clean EvaluationService instance.
func NewEvaluationService(dataSource EvaluationDataSource) EvaluationService {
	return &evaluationService{
		dataSource: dataSource,
	}
}

// Evaluate retrieves configuration from the data source and executes the evaluation algorithm.
func (s *evaluationService) Evaluate(ctx context.Context, sdkKey, featureName, userID, accountID, appName, evalContext string) (bool, error) {
	if sdkKey == "" {
		return false, domain.ErrMissingSDKKey
	}
	if featureName == "" {
		return false, domain.ErrMissingFeatureName
	}
	if accountID == "" {
		return false, domain.ErrMissingAccountID
	}
	if appName == "" {
		return false, domain.ErrMissingAppName
	}

	// 1. Parse evalContext if provided
	var attrs map[string]string
	if evalContext != "" {
		attrs = make(map[string]string)
		if err := json.Unmarshal([]byte(evalContext), &attrs); err != nil {
			return false, domain.ErrInvalidContextFormat
		}
	}

	// 2. Get config from the data source, no longer concerned with its internal implementation (cache, DB, singleflight)
	payload, err := s.dataSource.GetLeanConfig(ctx, sdkKey)
	if err != nil {
		return false, fmt.Errorf("failed to get config from data source: %w", err)
	}

	// 3. Execute pure evaluation logic using the fetched config
	return s.evaluateWithPayload(payload, featureName, userID, accountID, attrs)
}

// evaluateWithPayload is a pure computation function that uses pre-fetched optimized config for evaluation.
func (s *evaluationService) evaluateWithPayload(payload *cache.LeanEnvironmentConfig, featureName, userID, accountID string, attrs map[string]string) (bool, error) {
	// 1. Validate environment and tenant status from the config payload
	if !payload.IsActive {
		return false, domain.ErrEnvironmentInactive
	}
	if payload.TenantStatus != domain.TenantStatusActive {
		return false, domain.ErrTenantInactive
	}

	// 2. Look up the specific feature flag by name in the config payload (O(1) complexity)
	targetFlag, ok := payload.FeatureFlags[featureName]
	if !ok {
		return false, domain.ErrFeatureFlagNotFound
	}

	// 3. Execute core evaluation logic
	for _, rule := range targetFlag.Rules {
		// An empty AccountID string means this is a "user-only" rule
		accountMatch := rule.AccountID != "" && rule.AccountID == accountID
		userOnlyRule := rule.AccountID == ""

		if accountMatch {
			if len(rule.UserIDs) == 0 {
				return true, nil // Account-level rule matched, no specific user specified
			}
			for _, whitelistedUserID := range rule.UserIDs {
				if userID == whitelistedUserID {
					return true, nil // User is in the whitelist for this account
				}
			}
		} else if userOnlyRule {
			for _, whitelistedUserID := range rule.UserIDs {
				if userID == whitelistedUserID {
					return true, nil // User is in the general whitelist
				}
			}
		}
	}

	// 4. If no whitelist rule matches, check percentage rollout
	if targetFlag.RolloutPercentage > 0 {
		rolloutID := userID
		if rolloutID == "" {
			rolloutID = accountID
		}
		if rolloutID != "" {
			h := fnv.New32a()
			h.Write([]byte(rolloutID + ":" + featureName))
			bucket := h.Sum32() % 100
			if bucket < uint32(targetFlag.RolloutPercentage) {
				return true, nil
			}
		}
	}

	// 5. Evaluate segment conditions if user attributes are provided
	if len(targetFlag.Segments) > 0 && len(attrs) > 0 {
		for _, seg := range targetFlag.Segments {
			if evaluateSegment(seg, attrs) {
				return true, nil
			}
		}
	}

	// 6. No rule matches, not in rollout, no segment match — return false
	return false, nil
}

// evaluateSegment checks if the user attributes match a segment's conditions.
func evaluateSegment(seg cache.LeanSegment, attrs map[string]string) bool {
	if len(seg.Conditions) == 0 {
		return false
	}

	for _, cond := range seg.Conditions {
		attrVal, ok := attrs[cond.AttributeKey]
		if !ok {
			if seg.MatchType == "ALL" {
				return false
			}
			continue
		}

		matched := evaluateCondition(cond.Operator, attrVal, cond.Value)
		if seg.MatchType == "ALL" && !matched {
			return false
		}
		if seg.MatchType != "ALL" && matched {
			return true
		}
	}

	return seg.MatchType == "ALL"
}

// evaluateCondition checks a single condition against an attribute value.
func evaluateCondition(operator, attrVal, condVal string) bool {
	switch operator {
	case "equals":
		return attrVal == condVal
	case "not_equals":
		return attrVal != condVal
	case "contains":
		return strings.Contains(attrVal, condVal)
	case "not_contains":
		return !strings.Contains(attrVal, condVal)
	case "starts_with":
		return strings.HasPrefix(attrVal, condVal)
	case "ends_with":
		return strings.HasSuffix(attrVal, condVal)
	case "in":
		for _, item := range strings.Split(condVal, ",") {
			if attrVal == strings.TrimSpace(item) {
				return true
			}
		}
		return false
	case "not_in":
		for _, item := range strings.Split(condVal, ",") {
			if attrVal == strings.TrimSpace(item) {
				return false
			}
		}
		return true
	case "gt":
		return attrVal > condVal
	case "gte":
		return attrVal >= condVal
	case "lt":
		return attrVal < condVal
	case "lte":
		return attrVal <= condVal
	default:
		return false
	}
}
