package keys

import "strings"

// TODO(nsiow) write tests for the case-insensitivity of condition keys

// This const block holds string constants corresponding to AWS global condition keys
const (
	PrincipalTagPrefix = "aws:principaltag"
	ResourceTagPrefix  = "aws:resourcetag"
	RequestTagPrefix   = "aws:requesttag"

	PrincipalArn              = "aws:principalarn"
	PrincipalAccount          = "aws:principalaccount"
	PrincipalOrgPaths         = "aws:principalorgpaths"
	PrincipalOrgId            = "aws:principalorgid"
	PrincipalIsAwsService     = "aws:principalisawsservice"
	PrincipalServiceName      = "aws:principalservicename"
	PrincipalServiceNamesList = "aws:principalservicenameslist"
	PrincipalType             = "aws:principaltype"
	UserId                    = "aws:userid"
	Username                  = "aws:username"

	FederatedProvider            = "aws:federatedprovider"
	TokenIssueTime               = "aws:tokenissuetime"
	MultiFactorAuthAge           = "aws:multifactorauthage"
	MultiFactorAuthPresent       = "aws:multifactorauthpresent"
	Ec2InstanceSourceVpc         = "aws:ec2instancesourcevpc"
	Ec2InstanceSourcePrivateIPv4 = "aws:ec2instancesourceprivateipv4"
	SourceIdentity               = "aws:sourceidentity"
	RoleDelivery                 = "ec2:roledelivery"
	SourceInstanceArn            = "ec2:sourceinstancearn"

	SourceIp    = "aws:sourceip"
	SourceVpc   = "aws:sourcevpc"
	SourceVpce  = "aws:sourcevpce"
	VpcSourceIp = "aws:vpcsourceip"

	ResourceAccount  = "aws:resourceaccount"
	ResourceOrgPaths = "aws:resourceorgpaths"
	ResourceOrgId    = "aws:resourceorgid"

	CalledVia       = "aws:calledvia"
	CalledViaFirst  = "aws:calledviafirst"
	CalledViaLast   = "aws:calledvialast"
	ViaAwsService   = "aws:viaawsservice"
	CurrentTime     = "aws:currenttime"
	EpochTime       = "aws:epochtime"
	Referer         = "aws:referer"
	RequestedRegion = "aws:requestedregion"
	TagKeys         = "aws:tagkeys"
	SecureTransport = "aws:securetransport"
	SourceArn       = "aws:sourcearn"
	SourceAccount   = "aws:sourceaccount"
	SourceOrgPaths  = "aws:sourceorgpaths"
	SourceOrgId     = "aws:sourceorgid"
	UserAgent       = "aws:useragent"
)

var GLOBAL_COND_KEYS = map[string]struct{}{
	PrincipalArn:                 {},
	PrincipalAccount:             {},
	PrincipalOrgPaths:            {},
	PrincipalOrgId:               {},
	PrincipalIsAwsService:        {},
	PrincipalServiceName:         {},
	PrincipalServiceNamesList:    {},
	PrincipalType:                {},
	UserId:                       {},
	Username:                     {},
	FederatedProvider:            {},
	TokenIssueTime:               {},
	MultiFactorAuthAge:           {},
	MultiFactorAuthPresent:       {},
	Ec2InstanceSourceVpc:         {},
	Ec2InstanceSourcePrivateIPv4: {},
	SourceIdentity:               {},
	RoleDelivery:                 {},
	SourceInstanceArn:            {},
	SourceIp:                     {},
	SourceVpc:                    {},
	SourceVpce:                   {},
	VpcSourceIp:                  {},
	ResourceAccount:              {},
	ResourceOrgPaths:             {},
	ResourceOrgId:                {},
	CalledVia:                    {},
	CalledViaFirst:               {},
	CalledViaLast:                {},
	ViaAwsService:                {},
	CurrentTime:                  {},
	EpochTime:                    {},
	Referer:                      {},
	RequestedRegion:              {},
	TagKeys:                      {},
	SecureTransport:              {},
	SourceArn:                    {},
	SourceAccount:                {},
	SourceOrgPaths:               {},
	SourceOrgId:                  {},
	UserAgent:                    {},
}

func IsGlobalConditionKey(key string) bool {
	key = strings.ToLower(key)
	_, ok := GLOBAL_COND_KEYS[key]
	return ok
}
