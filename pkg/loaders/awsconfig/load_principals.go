package awsconfig

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

var principalTypes = []string{
	CONST_TYPE_AWS_IAM_ROLE,
	CONST_TYPE_AWS_IAM_USER,
}

// loadPrincipals takes a list of AWS Config items and extracts resources
func loadPrincipals(items []ConfigItem, mpm *PolicyMap) ([]entities.Principal, error) {
	var ps []entities.Principal

	// Iterate through our AWS Config items
	for _, i := range items {

		// Filter out only Principal types
		if !slices.Contains(principalTypes, i.Type) {
			continue
		}

		// Load the single principal
		p, err := loadPrincipal(i, mpm)
		if err != nil {
			return nil, err
		}

		ps = append(ps, *p)
	}

	return ps, nil
}

// loadPrincipal takes a single AWS Config item and returns a parsed principal object
func loadPrincipal(i ConfigItem, mpm *PolicyMap) (*entities.Principal, error) {
	// Construct basic fields
	p := entities.Principal{
		Type:    i.Type,
		Account: i.Account,
		Region:  i.Region,
		Arn:     i.Arn,
		Tags:    i.Tags,
	}

	// Extract both inline and managed policies
	// TODO(nsiow) Give these errors improved context similar to managed policies
	ip, err := extractInlinePolicies(i)
	if err != nil {
		return nil, err
	}
	p.InlinePolicies = ip
	mp, err := extractManagedPolicies(i, mpm)
	if err != nil {
		return nil, err
	}
	p.ManagedPolicies = mp

	return &p, nil
}

// extractInlinePolicies attempts to retrieve the direct Principal permissions, if supported
func extractInlinePolicies(i ConfigItem) ([]policy.Policy, error) {
	switch i.Type {
	case CONST_TYPE_AWS_IAM_ROLE:
		// Fetch configuration.rolePolicyList
		var rpl rolePolicyListFragment
		err := json.Unmarshal(i.Configuration, &rpl)
		if err != nil {
			return nil, err
		}

		// Iterate over fragments and decode policy documents
		var policies []policy.Policy
		for _, f := range rpl.RolePolicyList {
			p, err := decodePolicyString(f.PolicyDocument)
			if err != nil {
				return nil, err
			}
			policies = append(policies, p)
		}

		return policies, nil
	default:
		return nil, fmt.Errorf("extractInlinePolicies not supported for type: %s", i.Type)
	}
}

// extractManagedPolicies attempts to retrieve the Principal's managed permisions, if supported
func extractManagedPolicies(i ConfigItem, mpm *PolicyMap) ([]policy.Policy, error) {
	switch i.Type {
	case CONST_TYPE_AWS_IAM_ROLE:
		// Fetch configuration.attachedManagedPolicies
		var mpl managedPolicyListFragment
		err := json.Unmarshal(i.Configuration, &mpl)
		if err != nil {
			return nil, err
		}

		// Iterate over fragments and look up policies by ARN
		var policies []policy.Policy
		for _, f := range mpl.AttachedManagedPolicies {
			p, exists := mpm.Get(f.PolicyArn)
			if !exists {
				return nil, fmt.Errorf("managed policy '%s' not found in provided map", f.PolicyArn)
			}
			policies = append(policies, p)
		}

		return policies, nil
	default:
		return nil, fmt.Errorf("extractManagedPolicies not supported for type: %s", i.Type)
	}
}

// managedPolicyListFragment allows for unmarshalling of Managed Policy configuration blobs
type managedPolicyListFragment struct {
	AttachedManagedPolicies []managedPolicyFragment `json:"attachedManagedPolicies"`
}

// managedPolicyFragment allows for unmarshalling of Managed Policy configuration blobs
type managedPolicyFragment struct {
	PolicyArn  string `json:"policyArn"`
	PolicyName string `json:"policyName"`
}

// rolePolicyListFragment allows for unmarshalling of Inline Policy configuration blobs
type rolePolicyListFragment struct {
	RolePolicyList []inlinePolicyFragment `json:"rolePolicyList"`
}

// inlinePolicyFragment allows for unmarshalling of Inline Policy configuration blobs
type inlinePolicyFragment struct {
	PolicyDocument string `json:"policyDocument"`
	PolicyName     string `json:"policyName"`
}
