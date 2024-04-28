package awsconfig

import (
	"encoding/json"
	"fmt"
)

// policyItemConfiguration is a struct describing a configuration fragment for AWS::IAM::Policy
type policyItemConfiguration struct {
	PolicyVersionList []policyItemVersion `json:"policyVersionList"`
}

// Default attempts to retrieve the default version of this policy
func (p *policyItemConfiguration) Default() (policyItemVersion, error) {
	for _, piv := range p.PolicyVersionList {
		if piv.IsDefaultVersion {
			return piv, nil
		}
	}

	return policyItemVersion{}, fmt.Errorf("no valid default policy version found for: %v", p)
}

// policyItemVersion is a struct describing a configuration fragment for AWS::IAM::Policy
type policyItemVersion struct {
	VersionId        string `json:"versionId"`
	IsDefaultVersion bool   `json:"isDefaultVersion"`
	Document         string `json:"document"`
}

// loadPolicies takes a list of AWS Config items and extracts policies into a ManagedPolicyMap
func loadPolicies(items []Item) (*ManagedPolicyMap, error) {
	m := NewManagedPolicyMap()

	// Iterate through our AWS Config items
	for _, i := range items {

		// Filter out only `AWS::IAM::Policy`
		if i.Type != CONST_TYPE_AWS_IAM_POLICY {
			continue
		}

		// Parse policy configuration
		pic := policyItemConfiguration{}
		err := json.Unmarshal(i.Configuration, &pic)
		if err != nil {
			return nil, err
		}

		// Attempt to retrieve default version
		defaultVersion, err := pic.Default()
		if err != nil {
			return nil, err
		}

		// Attempt to decode our policy string; this _should_ always work
		policy, err := decodePolicyString(defaultVersion.Document)
		if err != nil {
			return nil, err
		}

		// Save the decoded policy into our map
		m.Add(i.Arn, policy)
	}

	return m, nil
}
