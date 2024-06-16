package awsconfig

import (
	"encoding/json"
	"fmt"

	"github.com/nsiow/yams/pkg/policy"
)

// loadPolicies takes a list of AWS Config items and extracts policies into a PolicyMap
func loadPolicies(items []ConfigItem) (*PolicyMap, error) {
	pm := NewPolicyMap()

	// Iterate through our AWS Config items

	// First pass, we only look at AWS::IAM::Policy
	for _, i := range items {
		if i.Type != CONST_TYPE_AWS_IAM_POLICY {
			continue
		}

		// Save the decoded policy into our map
		policies, err := loadPolicy(i)
		if err != nil {
			return nil, err
		}
		pm.Add(CONST_TYPE_AWS_IAM_POLICY, i.Arn, policies)
	}

	// Second pass, we look at AWS::IAM::Group, which requires the policies be loaded into the map
	// in order to fully process
	for _, i := range items {
		if i.Type != CONST_TYPE_AWS_IAM_GROUP {
			continue
		}

		// Save the decoded policy into our map
		policies, err := loadGroup(i, pm)
		if err != nil {
			return nil, err
		}
		pm.Add(CONST_TYPE_AWS_IAM_GROUP, i.Arn, policies)
	}

	return pm, nil
}

// --------------------------------------------------------------------------------
// IAM Policies
// --------------------------------------------------------------------------------

// loadPolicy takes a single AWS Config item and returns a parsed policy object
func loadPolicy(i ConfigItem) ([]policy.Policy, error) {
	// Parse policy configuration
	pf := policyFragment{}
	err := json.Unmarshal(i.Configuration, &pf)
	if err != nil {
		return nil, fmt.Errorf("unable to parse policy version list (%+v): %v", pf, err)
	}

	// Attempt to retrieve default version
	var policies []policy.Policy
	for _, pv := range pf.PolicyVersionList {
		if pv.IsDefaultVersion {
			policies = append(policies, policy.Policy(pv.Document))
			break
		}
	}
	if len(policies) == 0 {
		return nil, fmt.Errorf("unable to determine default version for (%+v): %v", pf, err)
	}

	return policies, nil
}

// policyFragment is a struct describing a configuration fragment for AWS::IAM::Policy
type policyFragment struct {
	PolicyVersionList []struct {
		VersionId        string        `json:"versionId"`
		IsDefaultVersion bool          `json:"isDefaultVersion"`
		Document         encodedPolicy `json:"document"`
	} `json:"policyVersionList"`
}

// --------------------------------------------------------------------------------
// IAM Groups
// --------------------------------------------------------------------------------

// loadGroup takes a single AWS Config item and returns a parsed group policy object
func loadGroup(i ConfigItem, pm *PolicyMap) ([]policy.Policy, error) {
	// Parse group configuration
	gf := groupFragment{}
	err := json.Unmarshal(i.Configuration, &gf)
	if err != nil {
		return nil, fmt.Errorf("unable to parse group (%+v): %v", gf, err)
	}

	groupPolicies := []policy.Policy{}

	// Add attached group policies
	for _, p := range gf.AttachedManagedPolicies {
		policies, ok := pm.Get(CONST_TYPE_AWS_IAM_POLICY, p.PolicyArn)
		if !ok {
			return nil, fmt.Errorf("unable to find policy definition for group policy: %s", p.PolicyArn)
		}

		for _, p2 := range policies {
			groupPolicies = append(groupPolicies, p2)
		}
	}

	// Add inline group policies
	for _, p := range gf.GroupPolicyList {
		groupPolicies = append(groupPolicies, policy.Policy(p.PolicyDocument))
	}

	return groupPolicies, nil
}

// groupFragment is a struct describing a configuration fragment for AWS::IAM::Group
type groupFragment struct {
	AttachedManagedPolicies []struct {
		PolicyArn  string `json:"policyArn"`
		PolicyName string `json:"policyName"`
	} `json:"attachedManagedPolicies"`
	GroupPolicyList []struct {
		PolicyName     string        `json:"policyName"`
		PolicyDocument encodedPolicy `json:"policyDocument"`
	} `json:"groupPolicyList"`
}
