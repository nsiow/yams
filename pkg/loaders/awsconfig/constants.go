package awsconfig

const (
	// Define constants for AWS principal types
	CONST_TYPE_AWS_IAM_GROUP  = "AWS::IAM::Group"
	CONST_TYPE_AWS_IAM_POLICY = "AWS::IAM::Policy"
	CONST_TYPE_AWS_IAM_ROLE   = "AWS::IAM::Role"
	CONST_TYPE_AWS_IAM_USER   = "AWS::IAM::User"

	// Define constants for AWS resource types with special considerations
	CONST_TYPE_AWS_DYNAMODB_TABLE = "AWS::DynamoDB::Table"
	CONST_TYPE_AWS_KMS_KEY        = "AWS::KMS::Key"
	CONST_TYPE_AWS_S3_BUCKET      = "AWS::S3::Bucket"
	CONST_TYPE_AWS_SNS_TOPIC      = "AWS::SNS::Topic"
	CONST_TYPE_AWS_SQS_QUEUE      = "AWS::SQS::Queue"

	// Define constants for custom types
	CONST_TYPE_YAMS_ORGANIZATIONS_ACCOUNT = "Yams::Organizations::Account"
	CONST_TYPE_YAMS_ORGANIZATIONS_POLICY  = "Yams::Organizations::Policy"
)
