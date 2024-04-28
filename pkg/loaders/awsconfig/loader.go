package awsconfig

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

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

// NewLoader provisions and returns a new `Loader` struct, ready to use
func NewLoader() *Loader {
	return &Loader{}
}

// Principals returns all principals loaded by the target loader
func (l *Loader) Principals() []entities.Principal {
	return l.principals
}

// Resources returns all resources loaded by the target loader
func (l *Loader) Resources() []entities.Resource {
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

	// Some buffer customization, since these JSON blobs can get big
	buf := make([]byte, 0, 64*1024)
	s.Buffer(buf, 1024*1024)

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
	// Load policies first (required to load principals)
	mp, err := loadPolicies(items)
	if err != nil {
		return fmt.Errorf("error loading managed policies: %v", err)
	}
	a.managedPolicies = mp

	// Load principals
	p, err := loadPrincipals(items, mp)
	if err != nil {
		return fmt.Errorf("error loading principals: %v", err)
	}
	a.principals = p

	// Load resources
	r, err := loadResources(items)
	if err != nil {
		return fmt.Errorf("error loading resources: %v", err)
	}
	a.resources = r

	return nil
}
