package entities

import "github.com/nsiow/yams/pkg/policy"

// Group defines the general shape of an AWS IAM group
type Group struct {
	// uv is a reverse pointer back to the containing universe
	uv *Universe `json:"-"`

	// Type refers to the AWS resource type of the Resource
	Type string

	// AccountId refers to the 12-digit AWS account ID where the Resource resides
	AccountId string

	// Arn refers to the Amazon Resource Name of the Resource
	Arn Arn

	// InlinePolicies refers to the inline (unattached) policies associated with the Group
	InlinePolicies []policy.Policy

	// AttachedPolicies refers to the managed policies associated with the Group
	AttachedPolicies []Arn
}
