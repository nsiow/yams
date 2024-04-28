package awsconfig

import (
	"fmt"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

var principalTypes = []string{
	CONST_TYPE_AWS_IAM_ROLE,
	CONST_TYPE_AWS_IAM_USER,
}

// loadPrincipals takes a list of AWS Config items and extracts resources
func loadPrincipals(items []Item, mpm *ManagedPolicyMap) ([]entities.Principal, error) {
	var ps []entities.Principal

	// Iterate through our AWS Config items
	for _, i := range items {

		// Construct basic fields
		p := entities.Principal{
			Type:    i.Type,
			Account: i.Account,
			Region:  i.Region,
			Arn:     i.Arn,
			Tags:    i.Tags,
		}

		// Extract both inline and managed policies
		ip, err := extractInlinePolicies(i)
		if err != nil {
			return nil, err
		}
		p.InlinePolicies = ip
		mp, err := extractManagedPolicies(i)
		if err != nil {
			return nil, err
		}
		p.ManagedPolicies = mp

		ps = append(ps, p)
	}

	return ps, nil
}

// extractInlinePolicies attempts to retrieve the direct Principal permissions, if supported
func extractInlinePolicies(i Item) ([]policy.Policy, error) {
	switch i.Type {
	case CONST_TYPE_AWS_IAM_ROLE:

	default:
		return nil, fmt.Errorf("extractInlinePolicies not supported for type: %s", i.Type)
	}
}

// extractManagedPolicies attempts to retrieve the Principal's managed permisions, if supported
func extractManagedPolicies(i Item) ([]policy.Policy, error) {
	switch i.Type {
	default:
		return nil, fmt.Errorf("extractManagedPolicies not supported for type: %s", i.Type)
	}
}

// managedPolicyFragment allows for unmarshalling of Managed Policy configuration blobs
type managedPolicyFragment struct {
	PolicyArn  string `json:"policyArn"`
	PolicyName string `json:"policyName"`
}

// inlinePolicyFragment allows for unmarshalling of Inline Policy configuration blobs
type inlinePolicyFragment struct {
	PolicyDocument string `json:"policyDocument"`
	PolicyName     string `json:"policyName"`
}
