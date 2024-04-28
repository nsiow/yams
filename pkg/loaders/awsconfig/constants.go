package awsconfig

const (
	// Define constants for AWS principal types

	CONST_TYPE_AWS_IAM_ROLE   = "AWS::IAM::Role"
	CONST_TYPE_AWS_IAM_USER   = "AWS::IAM::User"
	CONST_TYPE_AWS_IAM_POLICY = "AWS::IAM::Policy"

	// Define constants for AWS resource types with special considerations

	CONST_TYPE_AWS_S3_BUCKET             = "AWS::S3::Bucket"
	CONST_TYPE_AWS_SNS_TOPIC             = "AWS::SNS::Topic"
	CONST_TYPE_AWS_SQS_QUEUE             = "AWS::SQS::Queue"
	CONST_TYPE_AWS_DYNAMODB_TABLE        = "AWS::DynamoDB::Table"
	CONST_TYPE_AWS_DYNAMODB_GLOBAL_TABLE = "AWS::DynamoDB::GlobalTable"
)
