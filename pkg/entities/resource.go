package entities

import (
	"fmt"
	"strings"

	"github.com/nsiow/yams/pkg/policy"
)

// Resource defines the general shape of an AWS cloud resource
type Resource struct {
	// uv is a reverse pointer back to the containing universe
	uv *Universe `json:"-"`

	// Type refers to the AWS resource type of the Resource
	Type string

	// AccountId refers to the 12-digit AWS account ID where the Resource resides
	AccountId string

	// Region refers to the AWS region ID where the Resource resides
	Region string

	// Name refers to the friendly name of the Resource
	Name string

	// Arn refers to the Amazon Resource Name of the Resource
	Arn Arn

	// Tags refers to the AWS metadata tags attached to the Resource
	Tags []Tag `json:",omitzero"`

	// Policy refers to the resource policy associated with the Resource
	Policy policy.Policy
}

func (r *Resource) Key() string {
	return r.Arn
}

func (r *Resource) Repr() (any, error) {
	return r.Freeze()
}

// Service derives the AWS service name from the resource type in form AWS::<Service>::<Type>
func (r *Resource) Service() (string, error) {
	components := strings.Split(r.Type, "::")
	if len(components) != 3 {
		return "", fmt.Errorf("cannot determined service from malformed type: %s", r.Type)
	}

	return strings.ToLower(components[1]), nil
}

// SubResource returns the ARN of the specified subresource
func (r *Resource) SubResource(subpath string) (*Resource, error) {
	switch r.Type {
	case "AWS::S3::Bucket":
		return &Resource{
			uv:        r.uv,
			Arn:       strings.TrimRight(r.Arn, "/") + "/" + strings.TrimLeft(subpath, "/"),
			Type:      "AWS::S3::Bucket::Object",
			AccountId: r.AccountId,
			Region:    r.Region,
			Policy:    r.Policy,
			Tags:      nil, // tags do not propagate automatically
		}, nil
	default:
		return nil, fmt.Errorf("do not know how to create sub-resource for: %s", r.Type)
	}
}
