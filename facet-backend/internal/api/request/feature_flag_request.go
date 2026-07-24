package request

// ListFeatureFlagsRequest defines query parameters for listing feature flags
type ListFeatureFlagsRequest struct {
	Page            int    `form:"page,default=1"`
	PageSize        int    `form:"pageSize,default=10"`
	Keyword         string `form:"keyword"`
	IncludeArchived bool   `form:"includeArchived"`
}

// WhitelistRule defines the structure of a single whitelist rule when creating/updating a feature flag
type WhitelistRule struct {
	Id        string   `json:"id"`
	AccountID string   `json:"accountId"`
	UserIDs   []string `json:"userIds"`
}

// CreateFeatureFlagRequest defines the request body for creating a feature flag (updated)
type CreateFeatureFlagRequest struct {
	Name              string          `json:"name" binding:"required"`
	Description       string          `json:"description"`
	Enabled           bool            `json:"enabled"`
	FlagType          string          `json:"flagType" binding:"required"`
	RolloutPercentage int             `json:"rolloutPercentage"`
	ExpiresAt         string          `json:"expiresAt"`
	WhitelistRules    []WhitelistRule `json:"whitelistRules"`
	SegmentIDs        []string        `json:"segmentIds"`
}

// UpdateFeatureFlagRequest defines the request body for updating a feature flag (updated)
type UpdateFeatureFlagRequest struct {
	Description       *string         `json:"description"`
	Enabled           bool            `json:"enabled"`
	FlagType          string          `json:"flagType"`
	RolloutPercentage int             `json:"rolloutPercentage"`
	ExpiresAt         string          `json:"expiresAt"`
	WhitelistRules    []WhitelistRule `json:"whitelistRules"`
	SegmentIDs        []string        `json:"segmentIds"`
}
