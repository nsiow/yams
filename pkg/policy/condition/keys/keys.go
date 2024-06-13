package keys

// This const block holds string constants corresponding to AWS global condition keys
// TODO(nsiow) add comments + references
const (
	PrincipalTagPrefix = "aws:PrincipalTag/"
	ResourceTagPrefix  = "aws:ResourceTag/"
	RequestTagPrefix   = "aws:RequestTag/"

	PrincipalArn              = "aws:PrincipalArn"
	PrincipalAccount          = "aws:PrincipalAccount"
	PrincipalOrgPaths         = "aws:PrincipalOrgPaths"
	PrincipalOrgId            = "aws:PrincipalOrgID"
	PrincipalIsAwsService     = "aws:PrincipalIsAWSService"
	PrincipalServiceName      = "aws:PrincipalServiceName"
	PrincipalServiceNamesList = "aws:PrincipalServiceNamesList"
	PrincipalType             = "aws:PrincipalType"
	UserId                    = "aws:userid"
	Username                  = "aws:username"

	FederatedProvider            = "aws:FederatedProvider"
	TokenIssueTime               = "aws:TokenIssueTime"
	MultiFactorAuthAge           = "aws:MultiFactorAuthAge"
	MultiFactorAuthPresent       = "aws:MultiFactorAuthPresent"
	Ec2InstanceSourceVpc         = "aws:Ec2InstanceSourceVpc"
	Ec2InstanceSourcePrivateIPv4 = "aws:Ec2InstanceSourcePrivateIPv4"
	SourceIdentity               = "aws:SourceIdentity"
	RoleDelivery                 = "ec2:RoleDelivery"
	SourceInstanceArn            = "ec2:SourceInstanceArn"

	SourceIp    = "aws:SourceIp"
	SourceVpc   = "aws:SourceVpc"
	SourceVpce  = "aws:SourceVpce"
	VpcSourceIp = "aws:VpcSourceIp"

	ResourceAccount  = "aws:ResourceAccount"
	ResourceOrgPaths = "aws:ResourceOrgPaths"
	ResourceOrgId    = "aws:ResourceOrgID"

	CalledVia       = "aws:CalledVia"
	CalledViaFirst  = "aws:CalledViaFirst"
	CalledViaLast   = "aws:CalledViaLast"
	ViaAwsService   = "aws:ViaAWSService"
	CurrentTime     = "aws:CurrentTime"
	EpochTime       = "aws:EpochTime"
	Referer         = "aws:referer"
	RequestedRegion = "aws:RequestedRegion"
	TagKeys         = "aws:TagKeys"
	SecureTransport = "aws:SecureTransport"
	SourceArn       = "aws:SourceArn"
	SourceAccount   = "aws:SourceAccount"
	SourceOrgPaths  = "aws:SourceOrgPaths"
	SourceOrgId     = "aws:SourceOrgID"
	UserAgent       = "aws:UserAgent"
)
