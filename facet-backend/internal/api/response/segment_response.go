package response

import "time"

type SegmentResponse struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	MatchType   string                `json:"matchType"`
	Conditions  []SegmentConditionDTO `json:"conditions"`
	CreatedAt   time.Time             `json:"createdAt"`
	UpdatedAt   time.Time             `json:"updatedAt"`
}

type SegmentConditionDTO struct {
	ID           string `json:"id"`
	AttributeKey string `json:"attributeKey"`
	Operator     string `json:"operator"`
	Value        string `json:"value"`
}

type ListSegmentsResponse struct {
	Page       int64             `json:"page"`
	PageSize   int64             `json:"pageSize"`
	Total      int64             `json:"total"`
	TotalPages int64             `json:"totalPages"`
	Items      []SegmentResponse `json:"items"`
}
