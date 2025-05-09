package entities

// Account defines the general shape of an AWS account
type Account struct {
	// uv is a reverse pointer back to the containing universe
	uv *Universe `json:"-"`

	// AccountId refers to the 12-digit ID of this AWS account
	Id string

	// OrgId refers to the ID of the AWS Organizations org where the Account resides
	OrgId string

	// OrgPaths refers to the collection of org-paths containing the account
	// TODO(nsiow) implement this in the org crawler
	OrgPaths []string

	// FIXME(nsiow) implement these as OrgNodes so that we have node information on where SCPs live

	// SCPs refers to the Service Control Policies applied to the account
	// TODO(nsiow) add more support for the niche cases described in:
	// https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_scps_syntax.html
	SCPs [][]Arn

	// RCPs refers to the Resource Control Policies applied to the account
	RCPs [][]Arn
}
