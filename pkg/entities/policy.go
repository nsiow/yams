package entities

import "github.com/nsiow/yams/pkg/policy"

// Policy defines the general shape of an actualized IAM policy
//
// This is distinct from the [policy.Policy] type which is focused on the data/grammar of a policy,
// representing instead an "external" (non-inline) policy with an addressable ARN, such as a
// customer managed policy or an SCP
type Policy struct {
	// Type refers to the AWS resource type of the Policy
	Type string

	// AccountId refers to the 12-digit AWS account ID where the Policy resides
	AccountId string

	// Arn refers to the Amazon Resource Name of the Policy
	Arn string

	// Policy contains the actual Policy data
	Policy policy.Policy
}
