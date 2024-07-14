package entities

import (
	"github.com/nsiow/yams/pkg/policy"
)

// Principal defines the general shape of an AWS cloud principal
type Principal struct {
	// Type refers to the AWS resource type of the Principal
	Type string

	// OrgId refers to the ID of the AWS Organizations org where the Principal resides
	OrgId string

	// OrgPath refers to the path of the AWS Organizations OU where the Principal resides
	OrgPath string

	// Account refers to the 12-digit AWS account ID where the Principal resides
	Account string

	// Arn refers to the Amazon Resource Name of the Principal
	Arn string

	// Tags refers to the AWS metadata tags attached to the Principal
	Tags []Tag

	// InlinePolicies refers to the inline (unattached) policies associated with the Principal
	InlinePolicies []policy.Policy

	// AttachedPolicies refers to the managed policies associated with the Principal
	AttachedPolicies []policy.Policy

	// GroupPolicies refers to the group inline/unattached policies associated with the Principal
	GroupPolicies []policy.Policy

	// PermissionsBoundary refers to the policy set as the Principal's permissions boundary
	PermissionsBoundary policy.Policy
}
