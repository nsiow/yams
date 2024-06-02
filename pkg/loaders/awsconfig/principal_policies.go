package awsconfig

import (
	"encoding/json"
	"fmt"

	"github.com/nsiow/yams/pkg/policy"
)

var principalTypes = []string{
	CONST_TYPE_AWS_IAM_ROLE,
	CONST_TYPE_AWS_IAM_USER,
}

// --------------------------------------------------------------------------------
// Common
// --------------------------------------------------------------------------------

// extractInlinePolicies attempts to retrieve the direct Principal permissions, if supported
func extractInlinePolicies(i ConfigItem) ([]policy.Policy, error) {
	switch i.Type {
	case CONST_TYPE_AWS_IAM_ROLE:
		return extractInlineRolePolicies(i)
	default:
		return nil, fmt.Errorf("extractInlinePolicies not supported for type: %s", i.Type)
	}
}

// extractManagedPolicies attempts to retrieve the Principal's managed permisions, if supported
func extractManagedPolicies(i ConfigItem, mpm *PolicyMap) ([]policy.Policy, error) {
	switch i.Type {
	case CONST_TYPE_AWS_IAM_ROLE:
		return extractAttachedRolePolicies(i, mpm)
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

// --------------------------------------------------------------------------------
// AWS IAM Roles
// --------------------------------------------------------------------------------

// rolePolicyListFragment allows for unmarshalling of Inline Policy configuration blobs
type rolePolicyListFragment struct {
	RolePolicyList []inlinePolicyFragment `json:"rolePolicyList"`
}

// inlinePolicyFragment allows for unmarshalling of Inline Policy configuration blobs
type inlinePolicyFragment struct {
	PolicyDocument string `json:"policyDocument"`
	PolicyName     string `json:"policyName"`
}

// extractInlineRolePolicies defines how to retrieve policies from an IAM role
func extractInlineRolePolicies(i ConfigItem) ([]policy.Policy, error) {
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
}

// extractAttachedRolePolicies defines how to retrieve attached managed policies for an IAM role
func extractAttachedRolePolicies(i ConfigItem, mpm *PolicyMap) ([]policy.Policy, error) {
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
}
