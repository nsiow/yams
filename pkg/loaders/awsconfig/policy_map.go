package awsconfig

import (
	"fmt"
	"strings"

	"github.com/nsiow/yams/pkg/policy"
)

// PolicyMap contains a mapping from ID (Arn/GroupName) to policy for AWS-managed policies,
// customer-managed policies, and customer-managed group policies
type PolicyMap struct {
	mapping map[string][]policy.Policy
}

// NewPolicyMap creates and returns an initialized instance of PolicyMap
func NewPolicyMap() *PolicyMap {
	m := PolicyMap{}
	m.mapping = make(map[string][]policy.Policy)
	return &m
}

// Add creates a new mapping between the provided Arn and policy
func (m *PolicyMap) Add(pType, arn string, pstruct []policy.Policy) {
	arn = m.NormalizeArn(pType, arn)
	m.mapping[arn] = pstruct
}

// Get retrieves the requested policy by Arn, if it exists
func (m *PolicyMap) Get(pType, arn string) ([]policy.Policy, bool) {
	arn = m.NormalizeArn(pType, arn)
	val, ok := m.mapping[arn]
	return val, ok
}

// NormalizeArn updates the arn so that it is able to be stored/retrieve by callers with
// potential inconsistencies in how they are performing the lookups
func (m *PolicyMap) NormalizeArn(pType, arn string) string {
	switch pType {
	case CONST_TYPE_AWS_IAM_POLICY:
		return m.NormalizePolicyArn(arn)
	case CONST_TYPE_AWS_IAM_GROUP:
		return m.NormalizeGroupArn(arn)
	default:
		return arn
	}
}

// NormalizePolicyArn helps with some inconsistencies in policy naming
func (m *PolicyMap) NormalizePolicyArn(arn string) string {
	// Some historical vs current lambda policy naming
	if strings.Contains(arn, "AWSLambdaFullAccess") {
		arn = strings.ReplaceAll(arn, "AWSLambdaFullAccess", "AWSLambda_FullAccess")
	}
	if strings.Contains(arn, "AWSLambdaReadOnlyAccess") {
		arn = strings.ReplaceAll(arn, "AWSLambdaReadOnlyAccess", "AWSLambda_ReadOnlyAccess")
	}
	return arn
}

// NormalizeGroupArn updates the arn of an IAM group to a normalized version (ignoring path).
//
// While this reduces accuracy when returning ARNs back to the user, it allows us to retrieve and
// reference group policies from parts of the code which only have access to account ID + group
// name
//
// This is most relevant when parsing data from AWS APIs or Config, where group attachments on an
// IAM user are referenced by name rather than ARN
func (m *PolicyMap) NormalizeGroupArn(arn string) string {
	// Break ARN into components
	components := strings.Split(arn, "/")

	// If we don't have enough components to treat this as a pseudo-ARN, just return what we have
	if len(components) < 2 {
		return arn
	}

	// Otherwise, we just want the first and last pieces; sans path
	return fmt.Sprintf("%s/%s", components[0], components[len(components)-1])
}
