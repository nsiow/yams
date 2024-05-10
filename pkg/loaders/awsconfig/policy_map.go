package awsconfig

import (
	"strings"

	"github.com/nsiow/yams/pkg/policy"
)

const (
	// The ARN prefix which precedes all AWS-managed policies (is missing account ID)
	awsPolicyPrefix = "arn:aws:iam::aws:policy/"
)

// PolicyMap contains a mapping from ARN to policy for AWS- and customer-managed policies
type PolicyMap struct {
	pmap map[string]policy.Policy
}

// NewPolicyMap creates and returns an initialized instance of PolicyMap
func NewPolicyMap() *PolicyMap {
	m := PolicyMap{}
	m.pmap = make(map[string]policy.Policy)
	return &m
}

// Add creates a new mapping between the provided ARN and policy
func (m *PolicyMap) Add(arn string, pstruct policy.Policy) {
	arn = m.NormalizeArn(arn)
	m.pmap[arn] = pstruct
}

// Get retrieves the requested policy by ARN, if it exists
func (m *PolicyMap) Get(arn string) (policy.Policy, bool) {
	arn = m.NormalizeArn(arn)
	val, ok := m.pmap[arn]
	return val, ok
}

// TODO(nsiow) revisit this normalization with a larger test data set
// NormalizeArn updates the ARN to avoid cases where /[aws-]service-role/ paths are messed up
func (m *PolicyMap) NormalizeArn(arn string) string {
	// Only applies to AWS managed roles
	if !strings.HasPrefix(arn, awsPolicyPrefix) {
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
