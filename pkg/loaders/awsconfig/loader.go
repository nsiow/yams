package awsconfig

import (
	"bufio"
	"bytes"
	"encoding/json"

	"github.com/nsiow/yams/pkg/entities"
)

// Loader provides the ability to load resources/principals from AWS Config data
type Loader struct {
	// principals contains all cloud principals
	principals []entities.Principal

	// resources contains all cloud resources
	resources []entities.Resource

	// managedPolicies contains a map of policy ARN to policy
	managedPolicies *ManagedPolicyMap
}

// Principals returns all principals loaded by the target loader
func (l *Loader) Principals(data []byte) []entities.Principal {
	return l.principals
}

// Resources returns all resources loaded by the target loader
func (l *Loader) Resources(data []byte) []entities.Resource {
	return l.resources
}

// LoadJson loads data from a provided JSON array
func (a *Loader) LoadJson(data []byte) error {
	var items []Item
	json.Unmarshal(data, &items)
	return a.loadItems(items)
}

// LoadJson loads data from the provided newline-separate JSONL input
func (a *Loader) LoadJsonl(data []byte) error {
	r := bytes.NewReader(data)
	s := bufio.NewScanner(r)

	var items []Item
	for s.Scan() {
		// Read the next line
		b := s.Bytes()

		// Unmarshal into a single item
		var i Item
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
func (a *Loader) loadItems(items []Item) error {
	return nil
}
