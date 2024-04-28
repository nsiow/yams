package entities

import (
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/tag"
)

// Resource defines the general shape of an AWS cloud resource
type Resource struct {
	Type    string
	Account string
	Region  string
	Arn     string
	Policy  policy.Policy
	Tags    []tag.Tag
}
