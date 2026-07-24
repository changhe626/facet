package request

type ConditionInput struct {
	AttributeKey string `json:"attributeKey"`
	Operator     string `json:"operator"`
	Value        string `json:"value"`
}

type CreateSegmentRequest struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	MatchType   string           `json:"matchType"`
	Conditions  []ConditionInput `json:"conditions"`
}

type UpdateSegmentRequest struct {
	ID          string
	Name        string
	Description string           `json:"description"`
	MatchType   string           `json:"matchType"`
	Conditions  []ConditionInput `json:"conditions"`
}
