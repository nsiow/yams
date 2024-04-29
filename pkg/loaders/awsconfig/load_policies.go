package awsconfig

import (
	"encoding/json"
	"fmt"

	"github.com/nsiow/yams/pkg/policy"
)

// loadPolicies takes a list of AWS Config items and extracts policies into a ManagedPolicyMap
func loadPolicies(items []Item) (*ManagedPolicyMap, error) {
	m := NewManagedPolicyMap()

	// Iterate through our AWS Config items
	for _, i := range items {

		// Filter out only `AWS::IAM::Policy`
		if i.Type != CONST_TYPE_AWS_IAM_POLICY {
			continue
		}

		// Load the single policy
		policy, err := loadPolicy(i)
		if err != nil {
			return nil, err
		}

		// Save the decoded policy into our map
		m.Add(i.Arn, *policy)
	}

	return m, nil
}

// loadPolicy takes a single AWS Config item and returns a parsed policy object
func loadPolicy(i Item) (*policy.Policy, error) {
	// Parse policy configuration
	pic := policyFragment{}
	err := json.Unmarshal(i.Configuration, &pic)
	if err != nil {
		return nil, fmt.Errorf("unable to parse policy version list (%+v): %v", pic, err)
	}

	// Attempt to retrieve default version
	defaultVersion, err := pic.Default()
	if err != nil {
		return nil, fmt.Errorf("unable to determine default version for (%+v): %v", pic, err)
	}

	// Attempt to decode our policy string; this _should_ always work
	policy, err := decodePolicyString(defaultVersion.Document)
	if err != nil {
		return nil, fmt.Errorf("unable to decode policy string: %v", err)
	}

	return &policy, nil
}

// policyFragment is a struct describing a configuration fragment for AWS::IAM::Policy
type policyFragment struct {
	PolicyVersionList []policyVersionFragment `json:"policyVersionList"`
}

// Default attempts to retrieve the default version of this policy
func (p *policyFragment) Default() (policyVersionFragment, error) {
	for _, piv := range p.PolicyVersionList {
		if piv.IsDefaultVersion {
			return piv, nil
		}
	}

	return policyVersionFragment{}, fmt.Errorf("no valid default policy version found for: %v", p)
}

// policyVersionFragment is a struct describing a configuration fragment for AWS::IAM::Policy
type policyVersionFragment struct {
	VersionId        string `json:"versionId"`
	IsDefaultVersion bool   `json:"isDefaultVersion"`
	Document         string `json:"document"`
}
