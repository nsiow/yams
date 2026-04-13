package awsconfig

const (
	// AWS principal types
	CONST_TYPE_AWS_IAM_GROUP  = "AWS::IAM::Group"
	CONST_TYPE_AWS_IAM_POLICY = "AWS::IAM::Policy"
	CONST_TYPE_AWS_IAM_ROLE   = "AWS::IAM::Role"
	CONST_TYPE_AWS_IAM_USER   = "AWS::IAM::User"

	// AWS resource types
	CONST_TYPE_AWS_DYNAMODB_TABLE  = "AWS::DynamoDB::Table"
	CONST_TYPE_AWS_KMS_KEY         = "AWS::KMS::Key"
	CONST_TYPE_AWS_LAMBDA_FUNCTION = "AWS::Lambda::Function"
	CONST_TYPE_AWS_S3_BUCKET       = "AWS::S3::Bucket"
	CONST_TYPE_AWS_SNS_TOPIC       = "AWS::SNS::Topic"
	CONST_TYPE_AWS_SQS_QUEUE       = "AWS::SQS::Queue"
)

// OrgPrefix is the namespace prefix for custom organization types.
// Default is "Yams"; override via ldflags or --org-prefix flag.
var OrgPrefix = "Yams"

// Custom organization types, computed from OrgPrefix
var (
	CONST_TYPE_YAMS_ORGANIZATIONS_ACCOUNT string
	CONST_TYPE_YAMS_ORGANIZATIONS_SCP     string
	CONST_TYPE_YAMS_ORGANIZATIONS_RCP     string
)

func init() { RecomputeOrgConstants() }

// RecomputeOrgConstants rebuilds the custom type strings from OrgPrefix.
// Call after changing OrgPrefix at runtime.
func RecomputeOrgConstants() {
	CONST_TYPE_YAMS_ORGANIZATIONS_ACCOUNT = OrgPrefix + "::Organizations::Account"
	CONST_TYPE_YAMS_ORGANIZATIONS_SCP = OrgPrefix + "::Organizations::ServiceControlPolicy"
	CONST_TYPE_YAMS_ORGANIZATIONS_RCP = OrgPrefix + "::Organizations::ResourceControlPolicy"
}
