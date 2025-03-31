package entities

import (
	"github.com/nsiow/yams/pkg/policy"
)

// Account defines the general shape of an AWS account
type Account struct {
	// AccountId refers to the 12-digit ID of this AWS account
	Id string

	// OrgId refers to the ID of the AWS Organizations org where the Account resides
	OrgId string

	// OrgPaths refers to the collection of org-paths containing the account
	// FIXME(nsiow) implement this in the org crawler
	OrgPaths []string

	// SCPs refers to the Service Control Policies applied to the account
	// TODO(nsiow) add more support for the niche cases described in:
	// https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_scps_syntax.html
	SCPs [][]policy.Policy
}
