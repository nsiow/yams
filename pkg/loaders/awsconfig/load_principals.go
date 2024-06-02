package awsconfig

import (
	"slices"

	"github.com/nsiow/yams/pkg/entities"
)

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

		ps = append(ps, p)
	}

	return ps, nil
}

// loadPrincipal takes a single AWS Config item and returns a parsed principal object
func loadPrincipal(i ConfigItem, mpm *PolicyMap) (entities.Principal, error) {
	// Construct basic fields
	p := entities.Principal{
		Type:    i.Type,
		Account: i.Account,
		Arn:     i.Arn,
		Tags:    i.Tags,
	}

	// Extract both inline and managed policies
	// TODO(nsiow) Give these errors improved context similar to managed policies
	ip, err := extractInlinePolicies(i)
	if err != nil {
		return p, err
	}
	p.InlinePolicies = ip

	mp, err := extractManagedPolicies(i, mpm)
	if err != nil {
		return p, err
	}
	p.AttachedPolicies = mp

	return p, nil
}
