package awsconfig

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/nsiow/yams/pkg/policy"
)

// decodePolicyString attempts to retrieve a structured policy from an AWS-encoded blob
func decodePolicyString(policyString string) (policy.Policy, error) {
	p := policy.Policy{}

	// TODO(nsiow) revisit this error handling
	// Attempt decode
	escaped, _ := url.QueryUnescape(policyString)

	// If we decoded a nested JSON string, go again
	if strings.HasPrefix(escaped, "\"") && strings.HasSuffix(escaped, "\"") {
		err := json.Unmarshal([]byte(escaped), &escaped)
		if err != nil {
			return p, fmt.Errorf("error unwrapping nested JSON string: %v for input: %s", err, escaped)
		}
	}

	// Attempt JSON unmarshalling
	err := json.Unmarshal([]byte(escaped), &p)
	if err != nil {
		return p, fmt.Errorf("error converting decoded policy into struct: %v for input: %s", err, escaped)
	}

	return p, nil
}
