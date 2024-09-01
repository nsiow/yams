package awsconfig

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/nsiow/yams/pkg/aws/managedpolicies"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// Loader provides the ability to load resources/principals from AWS Config data
type Loader struct {
	// principals contains all cloud principals
	principals []entities.Principal

	// resources contains all cloud resources
	resources []entities.Resource
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

// Environment returns an Environment containing the loaded Principals + Resources
func (l *Loader) Environment() entities.Environment {
	return entities.Environment{
		Principals: l.Principals(),
		Resources:  l.Resources(),
	}
}

// LoadJson loads data from a provided JSON array
// TODO(nsiow) consider having this load from io.Reader instead
func (a *Loader) LoadJson(data []byte) error {
	var items []ConfigItem
	err := json.Unmarshal(data, &items)
	if err != nil {
		return fmt.Errorf("unable to load data as JSON: %v", err)
	}
	return a.loadItems(items)
}

// LoadJson loads data from the provided newline-separate JSONL input
// TODO(nsiow) consider having this load from io.Reader instead
func (a *Loader) LoadJsonl(data []byte) error {
	r := bytes.NewReader(data)
	s := bufio.NewScanner(r)

	// Some buffer customization, since these JSON blobs can get big
	buf := make([]byte, 0, 64*1024)
	s.Buffer(buf, 1024*1024)

	var items []ConfigItem
	for s.Scan() {
		// Read the next line; skip empty lines
		b := s.Bytes()
		if len(b) == 0 {
			continue
		}

		// Unmarshal into a single item
		var i ConfigItem
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
func (a *Loader) loadItems(items []ConfigItem) error {
	// Load control policies first
	cps, err := loadControlPolicies(items)
	if err != nil {
		return fmt.Errorf("error loading control policies: %v", err)
	}

	// Load policies (required to load principals)
	policies, err := loadPolicies(items)
	if err != nil {
		return fmt.Errorf("error loading managed policies: %v", err)
	}

	// Load AWS-managed policies into the managed policy map
	// (required because AWS Config does not report on them)
	for arn, pol := range managedpolicies.All() {
		policies.Add(CONST_TYPE_AWS_IAM_POLICY, arn, []policy.Policy{pol})
	}

	// Load principals
	principals, err := loadPrincipals(items, cps.SCPs, policies)
	if err != nil {
		return fmt.Errorf("error loading principals: %v", err)
	}
	a.principals = append(a.principals, principals...)

	// Load resources
	resources, err := loadResources(items)
	if err != nil {
		return fmt.Errorf("error loading resources: %v", err)
	}
	a.resources = append(a.resources, resources...)

	return nil
}
