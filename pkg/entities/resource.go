package entities

import (
	"github.com/nsiow/yams/pkg/policy"
)

// Resource defines the general shape of an AWS cloud resource
type Resource struct {
	Type    string
	Account string
	Region  string
	Arn     string
	Policy  policy.Policy
	Tags    []Tag
}
