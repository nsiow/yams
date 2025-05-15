package entities

// OrgNode defines the general shape of an Organizations Node: an Account or OU
type OrgNode struct {
	// Id refers to the primary identifier of the node
	Id string

	// Type refers to the type of the node, either ROOT or ORGANIZATIONAL_UNIT or ACCOUNT
	Type string

	// Arn refers to the Amazon Resource Name of the node
	Arn string

	// Name refers to the friendly name of the node; either account-name or ou-name
	Name string

	// SCPs refers to the Service Control Policies applied to the node
	// TODO(nsiow) add more support for the niche cases described in:
	// https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_policies_scps_syntax.html
	SCPs []Arn

	// RCPs refers to the Resource Control Policies applied to the node
	RCPs []Arn
}
