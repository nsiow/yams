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

	// If we find a nested JSON string, unmarshal it before continuing
	if strings.HasPrefix(policyString, `"`) {
		var inner string
		err := json.Unmarshal([]byte(policyString), &inner)
		if err != nil || len(inner) == 0 {
			return p, fmt.Errorf("error unwrapping nested JSON string: %v for input:\n%s", err, policyString)
		}
		policyString = inner
	}

	// Attempt unescaping
	escaped, err := url.QueryUnescape(policyString)
	if err != nil {
		return p, fmt.Errorf("error unescaping string: %v for input:\n%s", err, escaped)
	}

	// Attempt JSON unmarshalling
	err = json.Unmarshal([]byte(escaped), &p)
	if err != nil {
		return p, fmt.Errorf("error converting decoded policy into struct: %v for input:\n%s", err, escaped)
	}

	return p, nil
}
