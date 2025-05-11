package awsconfig

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/nsiow/yams/pkg/policy"
)

// EncodedPolicy is a (maybe URL encoded, maybe nested) string containing a policy document
type EncodedPolicy policy.Policy

// UnmarshalJSON instructs how to create EncodedPolicy fields from raw bytes
func (p *EncodedPolicy) UnmarshalJSON(data []byte) error {
	// Handle empty policy
	if len(data) == 0 {
		return nil
	}

	// Perform first unwrapping of string (if needed)
	var policyString string

	if data[0] == '"' {
		err := json.Unmarshal(data, &policyString)
		if err != nil {
			return fmt.Errorf("error in initial unwrapping of encoded policy (%v) for input %s", err, data)
		}

		// Empty string == empty policy
		if len(policyString) == 0 {
			*p = EncodedPolicy(policy.Policy{})
			return nil
		}
	} else {
		policyString = string(data)
	}

	// Attempt to decode
	policy, err := decodePolicyString(policyString)
	if err != nil {
		return fmt.Errorf("error decoding policy string: %w", err)
	}

	// Save to our policy
	*p = EncodedPolicy(policy)
	return nil
}

// decodePolicyString attempts to retrieve a structured policy from an AWS-encoded blob
func decodePolicyString(policyString string) (policy.Policy, error) {
	p := policy.Policy{}

	// If we find a nested JSON string, unmarshal it before continuing
	if strings.HasPrefix(policyString, `"`) {
		var inner string
		err := json.Unmarshal([]byte(policyString), &inner)
		if err != nil || len(inner) == 0 {
			return p, fmt.Errorf("error from nested JSON string: %v for input: %s", err, policyString)
		}
		policyString = inner
	}

	// Attempt unescaping
	escaped, err := url.QueryUnescape(policyString)
	if err != nil {
		return p, fmt.Errorf("error unescaping string: %v for input:\n%s", err, policyString)
	}

	// Attempt JSON unmarshalling
	err = json.Unmarshal([]byte(escaped), &p)
	if err != nil {
		return p, fmt.Errorf("error decoding policy into struct: %v for input: %s", err, escaped)
	}

	return p, nil
}
