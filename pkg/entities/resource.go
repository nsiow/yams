package entities

import (
	"fmt"
	"strings"

	"github.com/nsiow/yams/pkg/policy"
)

// Resource defines the general shape of an AWS cloud resource
type Resource struct {
	// Type refers to the AWS resource type of the Resource
	Type string

	// AccountId refers to the 12-digit AWS account ID where the Resource resides
	AccountId string

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
func (r *Resource) Service() (string, error) {
	components := strings.Split(r.Type, "::")
	if len(components) != 3 {
		return "", fmt.Errorf("cannot determined service from malformed type: %s", r.Type)
	}

	return strings.ToLower(components[1]), nil
}

// SubresourceArn returns the ARN of the specified subresource
func (r *Resource) SubresourceArn(subpath string) string {
	return strings.TrimRight(r.Arn, "/") + "/" + strings.TrimLeft(subpath, "/")
}
