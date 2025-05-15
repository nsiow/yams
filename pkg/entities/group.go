package entities

import "github.com/nsiow/yams/pkg/policy"

// Group defines the general shape of an AWS IAM group
type Group struct {
	// uv is a reverse pointer back to the containing universe
	uv *Universe `json:"-"`

	// Type refers to the AWS resource type of the Group
	Type string

	// AccountId refers to the 12-digit AWS account ID where the Group resides
	AccountId string

	// Name refers to the friendly name of the Group
	Name string

	// Arn refers to the Amazon Group Name of the Group
	Arn Arn

	// InlinePolicies refers to the inline (unattached) policies associated with the Group
	InlinePolicies []policy.Policy

	// AttachedPolicies refers to the managed policies associated with the Group
	AttachedPolicies []Arn
}

func (g *Group) Key() string {
	return g.Arn
}

func (g *Group) Repr() (any, error) {
	return g.Freeze()
}
