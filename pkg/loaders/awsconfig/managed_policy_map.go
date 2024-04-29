package awsconfig

import (
	"strings"

	"github.com/nsiow/yams/pkg/policy"
)

const (
	awsManagedPolicyPrefix = "arn:aws:iam::aws:policy/"
)

// ManagedPolicyMap contains a mapping from ARN to policy for AWS- and customer-managed policies
type ManagedPolicyMap struct {
	pmap map[string]policy.Policy
}

// NewManagedPolicyMap creates and returns an initialized instance of ManagedPolicyMap
func NewManagedPolicyMap() *ManagedPolicyMap {
	m := ManagedPolicyMap{}
	m.pmap = make(map[string]policy.Policy)
	return &m
}

// Add creates a new mapping between the provided ARN and policy
func (m *ManagedPolicyMap) Add(arn string, pstruct policy.Policy) {
	arn = m.NormalizeArn(arn)
	m.pmap[arn] = pstruct
}

// Get retrieves the requested policy by ARN, if it exists
func (m *ManagedPolicyMap) Get(arn string) (policy.Policy, bool) {
	arn = m.NormalizeArn(arn)
	val, ok := m.pmap[arn]
	return val, ok
}

// TODO(nsiow) revisit this normalization with a larger test data set
// NormalizeArn updates the ARN to avoid cases where /[aws-]service-role/ paths are messed up
func (m *ManagedPolicyMap) NormalizeArn(arn string) string {
	// Only applies to AWS managed roles
	if !strings.HasPrefix(arn, awsManagedPolicyPrefix) {
		return arn
	}

	// Perform a series of replacements to ensure normalization
	possibilities := []string{
		"aws:policy/aws-service-role/",
		"aws:policy/service-role/",
	}
	for _, p := range possibilities {
		arn = strings.ReplaceAll(arn, p, "aws:policy/")
	}

	return arn
}
