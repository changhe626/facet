package response

import (
	"facet-backend/internal/api/request"
	"facet-backend/internal/domain"
	"time"
)

// RuleSnapshot defines a minimal representation of a rule snapshot in history
type RuleSnapshot struct {
	AccountID string   `json:"accountId,omitempty"`
	UserIDs   []string `json:"userIds"`
}

// FeatureFlagListItem defines a single item in the feature flag list
type FeatureFlagListItem struct {
	ID                string      `json:"id"`
	Name              string      `json:"name"`
	Enabled           bool        `json:"enabled"`
	FlagType          string      `json:"flagType"`
	RolloutPercentage int         `json:"rolloutPercentage"`
	ExpiresAt         string      `json:"expiresAt"`
	ArchivedAt        *time.Time  `json:"archivedAt,omitempty"`
	UpdatedAt         time.Time   `json:"updatedAt"`
	UpdatedBy         string      `json:"updatedBy"`
}

// ListFeatureFlagsResponse defines the success response body for listing feature flags
type ListFeatureFlagsResponse struct {
	Page       int                   `json:"page"`
	PageSize   int                   `json:"pageSize"`
	Total      int64                 `json:"total"`
	TotalPages int                   `json:"totalPages"`
	Items      []FeatureFlagListItem `json:"items"`
}

// CreateFeatureFlagResponse defines the success response body for creating a feature flag
type CreateFeatureFlagResponse struct {
	ID                string                  `json:"id"`
	Name              string                  `json:"name"`
	Description       string                  `json:"description"`
	Enabled           bool                    `json:"enabled"`
	FlagType          string                  `json:"flagType"`
	RolloutPercentage int                     `json:"rolloutPercentage"`
	ExpiresAt         string                  `json:"expiresAt,omitempty"`
	WhitelistRules    []request.WhitelistRule `json:"whitelistRules"`
	CreatedAt         time.Time               `json:"createdAt"`
	CreatedBy         string                  `json:"createdBy"`
}

// FeatureFlagDetailResponse defines the success response body for getting feature flag details
type FeatureFlagDetailResponse struct {
	ID                string                  `json:"id"`
	Name              string                  `json:"name"`
	Description       string                  `json:"description"`
	Enabled           bool                    `json:"enabled"`
	FlagType          string                  `json:"flagType"`
	RolloutPercentage int                     `json:"rolloutPercentage"`
	ExpiresAt         string                  `json:"expiresAt,omitempty"`
	WhitelistRules    []request.WhitelistRule `json:"whitelistRules"`
	SegmentIDs        []string                `json:"segmentIds"`
	CreatedAt         time.Time               `json:"createdAt"`
	UpdatedAt         time.Time               `json:"updatedAt"`
	CreatedBy         string                  `json:"createdBy"`
	UpdatedBy         string                  `json:"updatedBy"`
}

// UpdateFeatureFlagResponse defines the success response body for updating a feature flag
type UpdateFeatureFlagResponse struct {
	ID                string                  `json:"id"`
	Name              string                  `json:"name"`
	Description       string                  `json:"description"`
	Enabled           bool                    `json:"enabled"`
	FlagType          string                  `json:"flagType"`
	RolloutPercentage int                     `json:"rolloutPercentage"`
	ExpiresAt         string                  `json:"expiresAt,omitempty"`
	WhitelistRules    []request.WhitelistRule `json:"whitelistRules"`
	CreatedAt         time.Time               `json:"createdAt"`
	UpdatedAt         time.Time               `json:"updatedAt"`
}

// FeatureFlagSnapshotResponse defines the API response structure for a single history version snapshot
type FeatureFlagSnapshotResponse struct {
	HistoryID      string            `json:"historyId"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	Enabled        bool              `json:"enabled"`
	FlagType       string            `json:"flagType"`
	ExpiresAt      string            `json:"expiresAt,omitempty"`
	WhitelistRules []RuleSnapshot    `json:"whitelistRules"`
	ChangeTime     time.Time         `json:"changeTime"`
	ChangeType     domain.ChangeType `json:"changeType"`
	ChangedBy      string            `json:"changedBy"`
}
