package entities

// ApiCall represents the authorization definition of a particular AWS API call
type ApiCall struct {
	Service          string   `json:"service"`
	Action           string   `json:"action"`
	AccessLevel      string   `json:"access_level"`
	Description      string   `json:"description"`
	ResourceTypes    []string `json:"resource_types"`
	ConditionKeys    []string `json:"condition_keys"`
	DependentActions []string `json:"dependent_actions"`
}
