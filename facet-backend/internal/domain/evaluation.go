package domain

import "errors"

// Evaluation-specific errors
var (
	ErrMissingSDKKey            = errors.New("sdkKey is required")
	ErrMissingFeatureName       = errors.New("featureName is required")
	ErrMissingAccountID         = errors.New("accountId is required")
	ErrMissingAppName           = errors.New("appName is required")
	ErrInvalidContextFormat     = errors.New("context must be a valid JSON object")
	ErrUnknownSegmentOperator   = errors.New("unknown segment condition operator")
	ErrInvalidInput             = errors.New("invalid input")
)
