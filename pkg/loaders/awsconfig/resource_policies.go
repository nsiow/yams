package awsconfig

import (
	"encoding/json"

	"github.com/nsiow/yams/pkg/policy"
)

// --------------------------------------------------------------------------------
// Common
// --------------------------------------------------------------------------------

// extractResourcePolicy uses resource-specific techniques to pull the policy from a Config item
func extractResourcePolicy(item ConfigItem) (policy.Policy, error) {
	switch item.Type {
	case CONST_TYPE_AWS_DYNAMODB_TABLE:
		return extractTablePolicy(item)
	case CONST_TYPE_AWS_S3_BUCKET:
		return extractBucketPolicy(item)
	case CONST_TYPE_AWS_SNS_TOPIC:
		return extractTopicPolicy(item)
	case CONST_TYPE_AWS_SQS_QUEUE:
		return extractQueuePolicy(item)
	case CONST_TYPE_AWS_IAM_ROLE:
		return extractRolePolicy(item)
	default:
		return policy.Policy{}, nil
	}
}

// --------------------------------------------------------------------------------
// AWS S3 Buckets
// --------------------------------------------------------------------------------

// awsS3Bucket defines the relevant structure of an AWS::S3::Bucket in Config
type awsS3Bucket struct {
	ConfigItem
	SupplementaryConfiguration struct {
		BucketPolicy struct {
			PolicyText encodedPolicy `json:"policyText"`
		}
	} `json:"supplementaryConfiguration"`
}

// extractBucketPolicy defines how to retrieve the resource policy
func extractBucketPolicy(item ConfigItem) (policy.Policy, error) {
	x := awsS3Bucket{}
	err := json.Unmarshal(item.SupplementaryConfiguration, &x.SupplementaryConfiguration)
	return policy.Policy(x.SupplementaryConfiguration.BucketPolicy.PolicyText), err
}

// --------------------------------------------------------------------------------
// AWS DynamoDB Tables
// --------------------------------------------------------------------------------

// awsDynamodbTable defines the relevant structure of an AWS::DynamoDB::Table in Config
type awsDynamodbTable struct{}

// extractTablePolicy defines how to retrieve the resource policy from an AWS::S3::Bucket
func extractTablePolicy(item ConfigItem) (policy.Policy, error) {
	return policy.Policy{}, nil // TODO(nsiow) update this when DynamoDB resource policies are in Config
}

// --------------------------------------------------------------------------------
// AWS SNS Topics
// --------------------------------------------------------------------------------

// AwsSnsTopic defines the relevant structure of an AWS::SNS::Topic in Config
type AwsSnsTopic struct {
	Configuration struct {
		Policy encodedPolicy
	} `json:"configuration"`
}

// extractTopicPolicy defines how to retrieve the resource policy
func extractTopicPolicy(item ConfigItem) (policy.Policy, error) {
	x := AwsSnsTopic{}
	err := json.Unmarshal(item.Configuration, &x.Configuration)
	return policy.Policy(x.Configuration.Policy), err
}

// --------------------------------------------------------------------------------
// AWS SQS Queues
// --------------------------------------------------------------------------------

// AwsSqsQueue defines the relevant structure of an AWS::SQS::Queue in Config
type AwsSqsQueue struct {
	Configuration struct {
		Policy encodedPolicy
	} `json:"configuration"`
}

// extractQueuePolicy defines how to retrieve the resource policy
func extractQueuePolicy(item ConfigItem) (policy.Policy, error) {
	x := AwsSqsQueue{}
	err := json.Unmarshal(item.Configuration, &x.Configuration)
	return policy.Policy(x.Configuration.Policy), err
}

// --------------------------------------------------------------------------------
// AWS IAM Roles
// --------------------------------------------------------------------------------

// AwsIamRole defines the relevant structure of an AWS::IAM::Role in Config
type AwsIamRole struct {
	Configuration struct {
		AssumeRolePolicyDocument encodedPolicy `json:"assumeRolePolicyDocument"`
	} `json:"configuration"`
}

// extractRolePolicy defines how to retrieve the resource policy
func extractRolePolicy(item ConfigItem) (policy.Policy, error) {
	x := AwsIamRole{}
	err := json.Unmarshal(item.Configuration, &x.Configuration)
	return policy.Policy(x.Configuration.AssumeRolePolicyDocument), err
}
