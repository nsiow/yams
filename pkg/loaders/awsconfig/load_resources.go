package awsconfig

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// loadResources takes a list of AWS Config items and extracts resources
func loadResources(items []ConfigItem) ([]entities.Resource, error) {
	var rs []entities.Resource

	// Iterate through our AWS Config items
	for _, i := range items {

		// TODO(nsiow) give similar treatment to errors for other entities
		// Load the single resource
		r, err := loadResource(i)
		if err != nil {
			return nil, fmt.Errorf("error loading resource '%s': %v", i.Arn, err)
		}

		rs = append(rs, *r)
	}

	return rs, nil
}

// loadResource takes a single AWS Config item and returns a parsed resource object
func loadResource(i ConfigItem) (*entities.Resource, error) {
	// Construct basic fields
	r := entities.Resource{
		Type:    i.Type,
		Account: i.Account,
		Region:  i.Region,
		Arn:     i.Arn,
		Tags:    i.Tags,
	}

	// Add policy where supported
	p, err := extractPolicy(i)
	if err != nil {
		return nil, err
	}
	if p != nil {
		r.Policy = *p
	}

	return &r, nil
}

// extractPolicy attempts to retrieve the resource policy, if supported
func extractPolicy(i ConfigItem) (*policy.Policy, error) {
	switch i.Type {
	case CONST_TYPE_AWS_DYNAMODB_TABLE:
		// TODO(nsiow) add logic once support lands in AWS Config
		return nil, nil
	case CONST_TYPE_AWS_S3_BUCKET:
		return policyFromConfiguration(i.SupplementaryConfiguration, "BucketPolicy.policyText")
	case CONST_TYPE_AWS_SNS_TOPIC, CONST_TYPE_AWS_SQS_QUEUE:
		return policyFromConfiguration(i.Configuration, "Policy")
	case CONST_TYPE_AWS_IAM_ROLE:
		return policyFromConfiguration(i.Configuration, "assumeRolePolicyDocument")
	// TODO(nsiow) implement KMS key policies
	default:
		return nil, nil
	}
}

// policyFromConfiguration is a helper function to fetch a policy out of a configuration blob
//
// Supports dot-notation in `key` for nested configurations
func policyFromConfiguration(conf json.RawMessage, key string) (*policy.Policy, error) {
	var policy json.RawMessage

	keyPath := strings.Split(key, ".")
	for i, path := range keyPath {
		// Convert from json.RawMessage -> map[string]json.RawMessage
		var m map[string]json.RawMessage
		err := json.Unmarshal(conf, &m)
		if err != nil {
			return nil, err
		}

		// TODO(nsiow) write a regression test per-resource for missing policy configuration
		// Retrieve the requested key from our new map
		cur, exists := m[path]
		if !exists {
			return nil, nil
		}

		// If we have reached the end of our key path, save our result back to our policy variable
		if i == len(keyPath)-1 {
			policy = cur
		}

		// Otherwise, set our value as the updated configuration and try again
		conf = cur
	}

	// Decode policy
	decoded, err := decodePolicyString(string(policy))
	if err != nil {
		return nil, fmt.Errorf("unable to decode policy: %v", err)
	}
	return &decoded, err
}
