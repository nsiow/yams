package entities

import (
	"github.com/nsiow/yams/pkg/policy"
)

// Account defines the general shape of an AWS account
type Account struct {
	// OrgId refers to the ID of the AWS Organizations org where the Principal resides
	// TODO(nsiow) these need to be filled out based on Organizations data
	OrgId string

	// OrgPath refers to the path of the AWS Organizations OU where the Principal resides
	// TODO(nsiow) add a test case for trailing /
	OrgPath string

	// OrgPaths refers to the collection of org-paths containing the account
	// FIXME(nsiow) implement this in the org crawler
	OrgPaths []string

	// SCPs refers to the Service Control Policies applied to the account
	// TODO(nsiow) add more support for the niche cases described in:
	// https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_scps_syntax.html
	SCPs [][]policy.Policy
}
