package awsconfig

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// loadResources takes a list of AWS Config items and extracts resources
func loadResources(items []Item) ([]entities.Resource, error) {
	var res []entities.Resource

	// Iterate through our AWS Config items
	for _, i := range items {

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
		r.Policy = *p

		res = append(res, r)
	}

	return res, nil
}

// extractPolicy attempts to retrieve the resource policy, if supported
func extractPolicy(i Item) (*policy.Policy, error) {
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
	case CONST_TYPE_AWS_KMS_KEY:
		panic("AWS::KMS::Key is not yet implemented")
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

		// Retrieve the requested key from our new map
		cur, exists := m[path]
		if !exists {
			return nil, fmt.Errorf("key '%s' not present in configuration", key)
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
		return nil, fmt.Errorf("unable to decode policy: %v", string(policy))
	}
	return &decoded, err
}
