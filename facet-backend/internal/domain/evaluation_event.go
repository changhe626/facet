package domain

import "time"

// EvaluationEvent defines the complete information to record for each evaluation event
type EvaluationEvent struct {
	EventID   string    `json:"eventId"`
	EventTime time.Time `json:"eventTime"`

	// Request context
	SDKKey      string `json:"sdkKey"`
	FeatureName string `json:"featureName"`
	AccountID   string `json:"accountId,omitempty"`
	UserID      string `json:"userId,omitempty"`
	AppName     string `json:"appName,omitempty"`

	// Evaluation result
	MatchedResult        bool   `json:"matchedResult"`
	EvaluationDurationMs int64  `json:"evaluationDurationMs"`
	ErrorMessage         string `json:"errorMessage,omitempty"`

	// Network info
	ClientIP  string `json:"clientIp,omitempty"`
	UserAgent string `json:"userAgent,omitempty"`
	Host      string `json:"host,omitempty"`
	Proto     string `json:"proto,omitempty"`
	Referer   string `json:"referer,omitempty"`
}
