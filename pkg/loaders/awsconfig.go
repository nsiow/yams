package loaders

import (
	"bufio"
	"bytes"
	"encoding/json"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/tag"
)

// AwsConfigLoader provides the ability to load resources/principals from AWS Config data
type AwsConfigLoader struct {
	// principals contains all cloud principals
	principals []entities.Principal

	// resources contains all cloud resources
	resources []entities.Resource

	// managedPolicies contains a map of policy ARN to policy
	managedPolicies map[string]policy.Policy
}

// Principals returns all principals loaded by the target loader
func (a *AwsConfigLoader) Principals(data []byte) []entities.Principal {
	return a.principals
}

// Resources returns all resources loaded by the target loader
func (a *AwsConfigLoader) Resources(data []byte) []entities.Resource {
	return a.resources
}

// LoadJson loads data from a provided JSON array
func (a *AwsConfigLoader) LoadJson(data []byte) error {
	var items []AwsConfigItem
	json.Unmarshal(data, &items)
	return a.loadItems(items)
}

// LoadJson loads data from the provided newline-separate JSONL input
func (a *AwsConfigLoader) LoadJsonl(data []byte) error {
	r := bytes.NewReader(data)
	s := bufio.NewScanner(r)

	var items []AwsConfigItem
	for s.Scan() {
		// Read the next line
		b := s.Bytes()

		// Unmarshal into a single item
		var i AwsConfigItem
		err := json.Unmarshal(b, &i)
		if err != nil {
			return err
		}

		// Add to running list of items
		items = append(items, i)
	}

	// If we encountered an error scanning, return it
	if err := s.Err(); err != nil {
		return err
	}

	// Proceed to load
	return a.loadItems(items)
}

// loadItems loads data from the provided AWS Config items
func (a *AwsConfigLoader) loadItems(items []AwsConfigItem) error {
	return nil
}

// AwsConfigItem defines the structure of a generic CI from AWS Config
type AwsConfigItem struct {
	Type                       string                     `json:"resourceType"`
	Account                    string                     `json:"accountId"`
	Region                     string                     `json:"awsRegion"`
	Arn                        string                     `json:"arn"`
	Tags                       []tag.Tag                  `json:"tags"`
	Configuration              map[string]json.RawMessage `json:"configuration"`
	SupplementaryConfiguration map[string]json.RawMessage `json:"supplementaryConfiguration"`
}
