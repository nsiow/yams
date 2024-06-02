package entities

import (
	"strings"

	"github.com/nsiow/yams/pkg/policy"
)

// Resource defines the general shape of an AWS cloud resource
type Resource struct {
	// Type refers to the AWS resource type of the Resource
	Type string

	// Account refers to the 12-digit AWS account ID where the Resource resides
	Account string

	// Region refers to the AWS region ID where the Resource resides
	Region string

	// Arn refers to the Amazon Resource Name of the Resource
	Arn string

	// Tags refers to the AWS metadata tags attached to the Resource
	Tags []Tag

	// Policy refers to the resource policy associated with the Resource
	Policy policy.Policy
}

// Service derives the AWS service name from the resource type in form AWS::<Service>::<Type>
func (r *Resource) Service() string {
	return strings.ToLower(strings.Split(r.Type, "::")[1])
}
