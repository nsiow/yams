package entities

import (
	"github.com/nsiow/yams/pkg/policy"
)

// Principal defines the general shape of an AWS IAM principal
type Principal struct {
	// Type refers to the AWS resource type of the Principal
	Type string

	// AccountId refers to the 12-digit AWS account ID where the Principal resides
	AccountId string

	// Arn refers to the Amazon Resource Name of the Principal
	Arn string

	// Tags refers to the AWS metadata tags attached to the Principal
	Tags []Tag

	// InlinePolicies refers to the inline (unattached) policies associated with the Principal
	InlinePolicies []policy.Policy

	// AttachedPolicies refers to the managed policies associated with the Principal
	AttachedPolicies []Arn

	// Groups refers to the AWS IAM groups to which the Principal belongs (only valid for
	// AWS::IAM::User types)
	Groups []Arn

	// PermissionsBoundary refers to the policy set as the Principal's permissions boundary
	PermissionsBoundary Arn
}
