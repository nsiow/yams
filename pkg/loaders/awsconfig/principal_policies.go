package awsconfig

import (
	"encoding/json"
	"fmt"

	"github.com/nsiow/yams/pkg/policy"
)

// --------------------------------------------------------------------------------
// Common
// --------------------------------------------------------------------------------

// extractInlinePolicies attempts to retrieve the direct Principal permissions, if supported
func extractInlinePolicies(i ConfigItem) ([]policy.Policy, error) {
	switch i.Type {
	case CONST_TYPE_AWS_IAM_ROLE:
		return extractInlineRolePolicies(i)
	case CONST_TYPE_AWS_IAM_USER:
		return extractInlineUserPolicies(i)
	default:
		return nil, fmt.Errorf("extractInlinePolicies not supported for type: %s", i.Type)
	}
}

// extractManagedPolicies attempts to retrieve the Principal's managed permisions, if supported
func extractManagedPolicies(i ConfigItem, pm *PolicyMap) ([]policy.Policy, error) {
	switch i.Type {
	case CONST_TYPE_AWS_IAM_ROLE:
		return extractAttachedRolePolicies(i, pm)
	case CONST_TYPE_AWS_IAM_USER:
		return extractAttachedUserPolicies(i, pm)
	default:
		return nil, fmt.Errorf("extractManagedPolicies not supported for type: %s", i.Type)
	}
}

// extractGroupPolicies attempts to retrieve the relevant Group permissions, if supported
func extractGroupPolicies(i ConfigItem, pm *PolicyMap) ([]policy.Policy, error) {
	switch i.Type {
	case CONST_TYPE_AWS_IAM_USER:
		return extractGroupUserPolicies(i, pm)
	default:
		return nil, fmt.Errorf("extractGroupInlinePolicies not supported for type: %s", i.Type)
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
	// Parse the relevant bits of the configuration
	var f rolePolicyListFragment
	err := json.Unmarshal(i.Configuration, &f)
	if err != nil {
		return nil, err
	}

	// Iterate over fragments and decode policy documents
	var policies []policy.Policy
	for _, f := range f.RolePolicyList {
		p, err := decodePolicyString(f.PolicyDocument)
		if err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}

	return policies, nil
}

// extractAttachedRolePolicies defines how to retrieve attached managed policies for an IAM role
func extractAttachedRolePolicies(i ConfigItem, pm *PolicyMap) ([]policy.Policy, error) {
	// Parse the relevant bits of the configuration
	var f managedPolicyListFragment
	err := json.Unmarshal(i.Configuration, &f)
	if err != nil {
		return nil, err
	}

	// Iterate over fragments and look up policies by ARN
	var policies []policy.Policy
	for _, f := range f.AttachedManagedPolicies {
		ps, exists := pm.Get(CONST_TYPE_AWS_IAM_POLICY, f.PolicyArn)
		if !exists {
			return nil, fmt.Errorf("managed policy '%s' not found in provided map", f.PolicyArn)
		}

		for _, p := range ps {
			policies = append(policies, p)
		}
	}

	return policies, nil
}

// --------------------------------------------------------------------------------
// AWS IAM Users
// --------------------------------------------------------------------------------

// userPolicyListFragment allows for unmarshalling of Inline Policy configuration blobs
type userPolicyListFragment struct {
	UserPolicyList []inlinePolicyFragment `json:"userPolicyList"`
}

// extractInlineUserPolicies defines how to retrieve inline policies from an IAM user
func extractInlineUserPolicies(i ConfigItem) ([]policy.Policy, error) {
	// Parse the relevant bits of the configuration
	var f userPolicyListFragment
	err := json.Unmarshal(i.Configuration, &f)
	if err != nil {
		return nil, err
	}

	// Iterate over fragments and decode policy documents
	var policies []policy.Policy
	for _, f := range f.UserPolicyList {
		p, err := decodePolicyString(f.PolicyDocument)
		if err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}

	return policies, nil
}

// extractAttachedUserPolicies defines how to retrieve attached managed policies for an IAM user
func extractAttachedUserPolicies(i ConfigItem, pm *PolicyMap) ([]policy.Policy, error) {
	// Parse the relevant bits of the configuration
	var f managedPolicyListFragment
	err := json.Unmarshal(i.Configuration, &f)
	if err != nil {
		return nil, err
	}

	// Iterate over fragments and look up policies by ARN
	var policies []policy.Policy
	for _, f := range f.AttachedManagedPolicies {
		ps, exists := pm.Get(CONST_TYPE_AWS_IAM_POLICY, f.PolicyArn)
		if !exists {
			return nil, fmt.Errorf("managed policy '%s' not found in provided map", f.PolicyArn)
		}

		for _, p := range ps {
			policies = append(policies, p)
		}
	}

	return policies, nil
}

// --------------------------------------------------------------------------------
// AWS IAM Groups
// --------------------------------------------------------------------------------

// groupListFragment allows for enumerating group memberships for an IAM user
type groupListFragment struct {
	GroupList []string `json:"groupList"`
}

// extractGroupUserPolicies defines how to retrieve attached + inline policies from an IAM group
func extractGroupUserPolicies(i ConfigItem, pm *PolicyMap) ([]policy.Policy, error) {
	// Parse the relevant bits of the configuration
	var f groupListFragment
	err := json.Unmarshal(i.Configuration, &f)
	if err != nil {
		return nil, err
	}

	// Perform group lookups and keep a running
	policies := []policy.Policy{}
	for _, groupName := range f.GroupList {
		groupArn := fmt.Sprintf("arn:aws:iam::%s:group/%s", i.Account, groupName)
		ps, exists := pm.Get(CONST_TYPE_AWS_IAM_GROUP, groupArn)
		if !exists {
			return nil, fmt.Errorf("group '%s' not found in provided map", groupArn)
		}

		policies = append(policies, ps...)
	}

	return policies, nil
}
