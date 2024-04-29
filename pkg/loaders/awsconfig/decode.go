package awsconfig

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/nsiow/yams/pkg/policy"
)

// decodePolicyString attempts to retrieve a structured policy from an AWS-encoded blob
func decodePolicyString(policyString string) (policy.Policy, error) {
	p := policy.Policy{}

	// Attempt decode
	escaped, err := url.QueryUnescape(policyString)
	if err != nil {
		return p, fmt.Errorf("unable to decode string '%s': %v", policyString, err)
	}

	// Attempt JSON unmarshalling
	err = json.Unmarshal([]byte(escaped), &p)
	if err != nil {
		return p, fmt.Errorf("error converting decoded policy into struct '%s': %v", escaped, err)
	}

	return p, nil
}
