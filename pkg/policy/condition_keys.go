package policy

// This const block holds string constants corresponding to AWS global condition keys
// TODO(nsiow) add comments + references
const (
	CONDKEY_AWS_PRINCIPAL_ARN                = "aws:PrincipalArn"
	CONDKEY_AWS_PRINCIPAL_ACCOUNT            = "aws:PrincipalAccount"
	CONDKEY_AWS_PRINCIPAL_ORG_PATHS          = "aws:PrincipalOrgPaths"
	CONDKEY_AWS_PRINCIPAL_ORG_ID             = "aws:PrincipalOrgID"
	CONDKEY_AWS_PRINCIPAL_TAG_PREFIX         = "aws:PrincipalTag"
	CONDKEY_AWS_PRINCIPAL_IS_AWS_SERVICE     = "aws:PrincipalIsAWSService"
	CONDKEY_AWS_PRINCIPAL_SERVICE_NAME       = "aws:PrincipalServiceName"
	CONDKEY_AWS_PRINCIPAL_SERVICE_NAMES_LIST = "aws:PrincipalServiceNamesList"
	CONDKEY_AWS_PRINCIPAL_TYPE               = "aws:PrincipalType"
	CONDKEY_AWS_PRINCIPAL_USERID             = "aws:userid"
	CONDKEY_AWS_PRINCIPAL_USERNAME           = "aws:username"

	CONDKEY_AWS_SESSION_FEDERATED_PROVIDER  = "aws:FederatedProvider"
	CONDKEY_AWS_SESSION_TOKEN_ISSUE_TIME    = "aws:TokenIssueTime"
	CONDKEY_AWS_SESSION_MFA_AGE             = "aws:MultiFactorAuthAge"
	CONDKEY_AWS_SESSION_MFA_PRESENT         = "aws:MultiFactorAuthPresent"
	CONDKEY_AWS_SESSION_SOURCE_VPC          = "aws:Ec2InstanceSourceVpc"
	CONDKEY_AWS_SESSION_SOURCE_IPv4         = "aws:Ec2InstanceSourcePrivateIPv4"
	CONDKEY_AWS_SESSION_SOURCE_IDENTITY     = "aws:SourceIdentity"
	CONDKEY_AWS_SESSION_ROLE_DELIVERY       = "ec2:RoleDelivery"
	CONDKEY_AWS_SESSION_SOURCE_INSTANCE_ARN = "ec2:SourceInstanceArn"

	CONDKEY_AWS_NETWORK_SOURCE_IP     = "aws:SourceIp"
	CONDKEY_AWS_NETWORK_SOURCE_VPC    = "aws:SourceVpc"
	CONDKEY_AWS_NETWORK_SOURCE_VPCE   = "aws:SourceVpce"
	CONDKEY_AWS_NETWORK_VPC_SOURCE_IP = "aws:VpcSourceIp"

	CONDKEY_AWS_RESOURCE_ACCOUNT    = "aws:ResourceAccount"
	CONDKEY_AWS_RESOURCE_ORG_PATHS  = "aws:ResourceOrgPaths"
	CONDKEY_AWS_RESOURCE_ORG_ID     = "aws:ResourceOrgID"
	CONDKEY_AWS_RESOURCE_TAG_PREFIX = "aws:ResourceTag/"

	CONDKEY_AWS_REQUEST_CALLED_VIA         = "aws:CalledVia"
	CONDKEY_AWS_REQUEST_CALLED_VIA_FIRST   = "aws:CalledViaFirst"
	CONDKEY_AWS_REQUEST_CALLED_VIA_LAST    = "aws:CalledViaLast"
	CONDKEY_AWS_REQUEST_VIA_AWS_SERVICE    = "aws:ViaAWSService"
	CONDKEY_AWS_REQUEST_CURRENT_TIME       = "aws:CurrentTime"
	CONDKEY_AWS_REQUEST_EPOCH_TIME         = "aws:EpochTime"
	CONDKEY_AWS_REQUEST_REFERER            = "aws:referer"
	CONDKEY_AWS_REQUEST_REQUESTED_REGION   = "aws:RequestedRegion"
	CONDKEY_AWS_REQUEST_REQUEST_TAG_PREFIX = "aws:RequestTag/"
	CONDKEY_AWS_REQUEST_TAG_KEYS           = "aws:TagKeys"
	CONDKEY_AWS_REQUEST_SECURE_TRANSPORT   = "aws:SecureTransport"
	CONDKEY_AWS_REQUEST_SOURCE_ARN         = "aws:SourceArn"
	CONDKEY_AWS_REQUEST_SOURCE_ACCOUNT     = "aws:SourceAccount"
	CONDKEY_AWS_REQUEST_SOURCE_ORG_PATHS   = "aws:SourceOrgPaths"
	CONDKEY_AWS_REQUEST_SOURCE_ORG_ID      = "aws:SourceOrgID"
	CONDKEY_AWS_REQUEST_USER_AGENT         = "aws:UserAgent"
)
