package sim

import (
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/policy/condition"
)

// AuthContext defines the tertiary context of a request that can be used for authz decisions
type AuthContext struct {
	Action     string
	Principal  *entities.Principal
	Resource   *entities.Resource
	Properties map[string]policy.Value
}

func (ac *AuthContext) Key(key string) string {
	switch key {
	// case condition.Key_AwsPrincipalTagPrefix:
	// 	fallthrough
	// case condition.Key_AwsResourceTagPrefix:
	// 	fallthrough
	// case condition.Key_AwsRequestTagPrefix:
	// 	fallthrough
	// case condition.Key_AwsPrincipalArn:
	// 	fallthrough
	case condition.Key_AwsPrincipalAccount:
		return ac.Principal.Account
	// case condition.Key_AwsPrincipalOrgPaths:
	// 	fallthrough
	// case condition.Key_AwsPrincipalOrgId:
	// 	fallthrough
	// case condition.Key_AwsPrincipalIsAwsService:
	// 	fallthrough
	// case condition.Key_AwsPrincipalServiceName:
	// 	fallthrough
	// case condition.Key_AwsPrincipalServiceNamesList:
	// 	fallthrough
	// case condition.Key_AwsPrincipalType:
	// 	fallthrough
	// case condition.Key_AwsPrincipalUserId:
	// 	fallthrough
	// case condition.Key_AwsPrincipalUsername:
	// 	fallthrough
	// case condition.Key_AwsSessionFederatedProvider:
	// 	fallthrough
	// case condition.Key_AwsSessionTokenIssueTime:
	// 	fallthrough
	// case condition.Key_AwsSessionMfaAge:
	// 	fallthrough
	// case condition.Key_AwsSessionMfaPresent:
	// 	fallthrough
	// case condition.Key_AwsSessionSourceVpc:
	// 	fallthrough
	// case condition.Key_AwsSessionSourceIpv4:
	// 	fallthrough
	// case condition.Key_AwsSessionSourceIdentity:
	// 	fallthrough
	// case condition.Key_AwsSessionRoleDelivery:
	// 	fallthrough
	// case condition.Key_AwsSessionSourceInstanceArn:
	// 	fallthrough
	// case condition.Key_AwsNetworkSourceIp:
	// 	fallthrough
	// case condition.Key_AwsNetworkSourceVpc:
	// 	fallthrough
	// case condition.Key_AwsNetworkSourceVpce:
	// 	fallthrough
	// case condition.Key_AwsNetworkVpcSourceIp:
	// 	fallthrough
	case condition.Key_AwsResourceAccount:
		return ac.Resource.Account
	// case condition.Key_AwsResourceOrgPaths:
	// 	fallthrough
	// case condition.Key_AwsResourceOrgId:
	// 	fallthrough
	// case condition.Key_AwsRequestCalledVia:
	// 	fallthrough
	// case condition.Key_AwsRequestCalledViaFirst:
	// 	fallthrough
	// case condition.Key_AwsRequestCalledViaLast:
	// 	fallthrough
	// case condition.Key_AwsRequestViaAwsService:
	// 	fallthrough
	// case condition.Key_AwsRequestCurrentTime:
	// 	fallthrough
	// case condition.Key_AwsRequestEpochTime:
	// 	fallthrough
	// case condition.Key_AwsRequestReferer:
	// 	fallthrough
	// case condition.Key_AwsRequestRequestedRegion:
	// 	fallthrough
	// case condition.Key_AwsRequestTagKeys:
	// 	fallthrough
	// case condition.Key_AwsRequestSecureTransport:
	// 	fallthrough
	// case condition.Key_AwsRequestSourceArn:
	// 	fallthrough
	// case condition.Key_AwsRequestSourceAccount:
	// 	fallthrough
	// case condition.Key_AwsRequestSourceOrgPaths:
	// 	fallthrough
	// case condition.Key_AwsRequestSourceOrgId:
	// 	fallthrough
	// case condition.Key_AwsRequestUserAgent:
	// 	fallthrough
	default:
		return ""
	}
}
