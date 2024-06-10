package condition

// This const block holds string constants corresponding to AWS global condition keys
// TODO(nsiow) add comments + references
const (
	Key_AwsPrincipalTagPrefix = "aws:PrincipalTag/"
	Key_AwsResourceTagPrefix  = "aws:ResourceTag/"
	Key_AwsRequestTagPrefix   = "aws:RequestTag/"

	Key_AwsPrincipalArn              = "aws:PrincipalArn"
	Key_AwsPrincipalAccount          = "aws:PrincipalAccount"
	Key_AwsPrincipalOrgPaths         = "aws:PrincipalOrgPaths"
	Key_AwsPrincipalOrgId            = "aws:PrincipalOrgID"
	Key_AwsPrincipalIsAwsService     = "aws:PrincipalIsAWSService"
	Key_AwsPrincipalServiceName      = "aws:PrincipalServiceName"
	Key_AwsPrincipalServiceNamesList = "aws:PrincipalServiceNamesList"
	Key_AwsPrincipalType             = "aws:PrincipalType"
	Key_AwsPrincipalUserId           = "aws:userid"
	Key_AwsPrincipalUsername         = "aws:username"

	Key_AwsSessionFederatedProvider = "aws:FederatedProvider"
	Key_AwsSessionTokenIssueTime    = "aws:TokenIssueTime"
	Key_AwsSessionMfaAge            = "aws:MultiFactorAuthAge"
	Key_AwsSessionMfaPresent        = "aws:MultiFactorAuthPresent"
	Key_AwsSessionSourceVpc         = "aws:Ec2InstanceSourceVpc"
	Key_AwsSessionSourceIpv4        = "aws:Ec2InstanceSourcePrivateIPv4"
	Key_AwsSessionSourceIdentity    = "aws:SourceIdentity"
	Key_AwsSessionRoleDelivery      = "ec2:RoleDelivery"
	Key_AwsSessionSourceInstanceArn = "ec2:SourceInstanceArn"

	Key_AwsNetworkSourceIp    = "aws:SourceIp"
	Key_AwsNetworkSourceVpc   = "aws:SourceVpc"
	Key_AwsNetworkSourceVpce  = "aws:SourceVpce"
	Key_AwsNetworkVpcSourceIp = "aws:VpcSourceIp"

	Key_AwsResourceAccount  = "aws:ResourceAccount"
	Key_AwsResourceOrgPaths = "aws:ResourceOrgPaths"
	Key_AwsResourceOrgId    = "aws:ResourceOrgID"

	Key_AwsRequestCalledVia       = "aws:CalledVia"
	Key_AwsRequestCalledViaFirst  = "aws:CalledViaFirst"
	Key_AwsRequestCalledViaLast   = "aws:CalledViaLast"
	Key_AwsRequestViaAwsService   = "aws:ViaAWSService"
	Key_AwsRequestCurrentTime     = "aws:CurrentTime"
	Key_AwsRequestEpochTime       = "aws:EpochTime"
	Key_AwsRequestReferer         = "aws:referer"
	Key_AwsRequestRequestedRegion = "aws:RequestedRegion"
	Key_AwsRequestTagKeys         = "aws:TagKeys"
	Key_AwsRequestSecureTransport = "aws:SecureTransport"
	Key_AwsRequestSourceArn       = "aws:SourceArn"
	Key_AwsRequestSourceAccount   = "aws:SourceAccount"
	Key_AwsRequestSourceOrgPaths  = "aws:SourceOrgPaths"
	Key_AwsRequestSourceOrgId     = "aws:SourceOrgID"
	Key_AwsRequestUserAgent       = "aws:UserAgent"
)
