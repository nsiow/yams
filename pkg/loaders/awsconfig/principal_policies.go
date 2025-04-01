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

// extractPermissionsBoundary attempts to retrieve the Principal's perm boundary, if supported
func extractPermissionsBoundary(i ConfigItem, pm *PolicyMap) (policy.Policy, error) {
	switch i.Type {
	case CONST_TYPE_AWS_IAM_ROLE:
		return extractRolePermissionsBoundary(i, pm)
	case CONST_TYPE_AWS_IAM_USER:
		return extractUserPermissionsBoundary(i, pm)
	default:
		return policy.Policy{},
			fmt.Errorf("extractPermissionsBoundary not supported for type: %s", i.Type)
	}
}

// extractGroupPolicies attempts to retrieve the relevant Group permissions, if supported
func extractGroupPolicies(i ConfigItem, pm *PolicyMap) ([]policy.Policy, error) {
	return extractGroupUserPolicies(i, pm)
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
	RolePolicyList []rolePolicyInlineFragment `json:"rolePolicyList"`
}

// rolePolicyInlineFragment allows for unmarshalling of Inline Policy configuration blobs
type rolePolicyInlineFragment struct {
	PolicyDocument string `json:"policyDocument"`
	PolicyName     string `json:"policyName"`
}

// rolePolicyPermissionsBoundaryFragment allows for unmarshalling of perm boundary blobs
type rolePolicyPermissionsBoundaryFragment struct {
	PermissionsBoundary struct {
		Arn  string `json:"permissionsBoundaryArn"`
		Type string `json:"permissionsBoundaryType"`
	} `json:"permissionsBoundary"`
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

		policies = append(policies, ps...)
	}

	return policies, nil
}

// extractRolePermissionsBoundary defines how to retrieve a perm boundary for an IAM role
func extractRolePermissionsBoundary(i ConfigItem, pm *PolicyMap) (policy.Policy, error) {
	// Parse the relevant bits of the configuration
	var f rolePolicyPermissionsBoundaryFragment
	err := json.Unmarshal(i.Configuration, &f)
	if err != nil {
		return policy.Policy{}, err
	}

	// Look up policy by ARN
	arn := f.PermissionsBoundary.Arn

	// Handle empty case
	if len(arn) == 0 {
		return policy.Policy{}, nil
	}

	// ... otherwise resolve policy
	policyList, exists := pm.Get(CONST_TYPE_AWS_IAM_POLICY, arn)
	if !exists || len(policyList) == 0 {
		return policy.Policy{}, fmt.Errorf("boundary policy '%s' not found in provided map", arn)
	}

	return policyList[0], nil
}

// --------------------------------------------------------------------------------
// AWS IAM Users
// --------------------------------------------------------------------------------

// userPolicyListFragment allows for unmarshalling of Inline Policy configuration blobs
type userPolicyListFragment struct {
	UserPolicyList []rolePolicyInlineFragment `json:"userPolicyList"`
}

// userPolicyPermissionsBoundaryFragment allows for unmarshalling of perm boundary blobs
type userPolicyPermissionsBoundaryFragment struct {
	PermissionsBoundary struct {
		Arn  string `json:"permissionsBoundaryArn"`
		Type string `json:"permissionsBoundaryType"`
	} `json:"permissionsBoundary"`
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

		policies = append(policies, ps...)
	}

	return policies, nil
}

// extractUserPermissionsBoundary defines how to retrieve a perm boundary for an IAM user
func extractUserPermissionsBoundary(i ConfigItem, pm *PolicyMap) (policy.Policy, error) {
	// Parse the relevant bits of the configuration
	var f userPolicyPermissionsBoundaryFragment
	err := json.Unmarshal(i.Configuration, &f)
	if err != nil {
		return policy.Policy{}, err
	}

	// Look up policy by ARN
	arn := f.PermissionsBoundary.Arn

	// Handle empty case
	if len(arn) == 0 {
		return policy.Policy{}, nil
	}

	// ... otherwise resolve policy
	policyList, exists := pm.Get(CONST_TYPE_AWS_IAM_POLICY, arn)
	if !exists || len(policyList) == 0 {
		return policy.Policy{}, fmt.Errorf("boundary policy '%s' not found in provided map", arn)
	}

	return policyList[0], nil
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
	var policies []policy.Policy
	for _, groupName := range f.GroupList {
		groupArn := fmt.Sprintf("arn:aws:iam::%s:group/%s", i.AccountId, groupName)
		ps, exists := pm.Get(CONST_TYPE_AWS_IAM_GROUP, groupArn)
		if !exists {
			return nil, fmt.Errorf("group '%s' not found in provided map", groupArn)
		}

		policies = append(policies, ps...)
	}

	return policies, nil
}
