package awsconfig

import (
	"fmt"

	"github.com/nsiow/yams/pkg/entities"
)

// loadPrincipals takes a list of AWS Config items and extracts resources
func loadPrincipals(
	items []ConfigItem,
	scps *ControlPolicyMap,
	pm *PolicyMap,
) ([]entities.Principal, error) {
	var ps []entities.Principal

	// Iterate through our AWS Config items
	for _, i := range items {

		// Filter out only Principal types
		if i.Type != CONST_TYPE_AWS_IAM_ROLE && i.Type != CONST_TYPE_AWS_IAM_USER {
			continue
		}

		// Load the principal
		p, err := loadPrincipal(i, scps, pm)
		if err != nil {
			return nil, err
		}

		ps = append(ps, p)
	}

	return ps, nil
}

// loadPrincipal takes a single AWS Config item and returns a parsed principal object
func loadPrincipal(
	i ConfigItem,
	scps *ControlPolicyMap,
	pm *PolicyMap,
) (entities.Principal, error) {
	// Construct basic fields
	p := entities.Principal{
		Type:    i.Type,
		Account: i.Account,
		Arn:     i.Arn,
		Tags:    i.Tags,
	}

	// Extract both inline and managed policies
	ip, err := extractInlinePolicies(i)
	if err != nil {
		return p, fmt.Errorf("error extracting inline policies for '%s': %w", i.Arn, err)
	}
	p.InlinePolicies = ip

	mp, err := extractManagedPolicies(i, pm)
	if err != nil {
		return p, fmt.Errorf("error extracting attached policies for '%s': %w", i.Arn, err)
	}
	p.AttachedPolicies = mp

	// If we are dealing with an IAM user, load its group policies as well
	if i.Type == CONST_TYPE_AWS_IAM_USER {
		gp, err := extractGroupPolicies(i, pm)
		if err != nil {
			return p, fmt.Errorf("error extracting group policies for '%s': %w", i.Arn, err)
		}
		p.GroupPolicies = gp
	}

	// Load permissions boundary
	pb, err := extractPermissionsBoundary(i, pm)
	if err != nil {
		return p, fmt.Errorf("error extracting permissions boundary for '%s': %w", i.Arn, err)
	}
	p.PermissionsBoundary = pb

	// Load SCPs
	if scp, exists := scps.Get(p.Account); exists {
		p.SCPs = scp
	}

	return p, nil
}
