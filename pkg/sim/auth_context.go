package sim

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy/condition"
)

// TODO(nsiow) check condition key reference for single vs multi values

// AuthContext defines the tertiary context of a request that can be used for authz decisions
type AuthContext struct {
	Action    string
	Principal *entities.Principal
	Resource  *entities.Resource
	// TODO(nsiow) figure out if we want to keep string values or allow for multivalues, etc
	Properties map[string]string
}

// Static values
const (
	TIME_FORMAT = "2006-01-02T15:04:05"

	EMPTY = ""
	TRUE  = "true"
	FALSE = "false"
)

// VariableExpansionRegex defines the variable to use for expanding policy variables
var VariableExpansionRegex = regexp.MustCompile(`\${[a-zA-Z0-9]+:[a-zA-Z0-9]+}`)

func (ac *AuthContext) Key(key string) string {
	switch {
	case strings.HasPrefix(key, condition.Key_AwsPrincipalTagPrefix):
		return ac.extractTag(key, ac.Principal.Tags)
	case strings.HasPrefix(key, condition.Key_AwsResourceTagPrefix):
		return ac.extractTag(key, ac.Resource.Tags)
	case strings.HasPrefix(key, condition.Key_AwsRequestTagPrefix):
		break // fall back to normal property extraction behavior
	}

	switch key {
	// Tag prefixes

	// case condition.Key_AwsRequestTagPrefix:
	// 	fallthrough

	// Static keys

	case condition.Key_AwsPrincipalArn:
		return ac.Principal.Arn
	case condition.Key_AwsPrincipalAccount:
		return ac.Principal.Account
	// case condition.Key_AwsPrincipalOrgPaths:
	// 	fallthrough
	// case condition.Key_AwsPrincipalOrgId:
	// 	fallthrough
	case condition.Key_AwsPrincipalIsAwsService:
		return FALSE // we only model IAM entities; never services
	case condition.Key_AwsPrincipalServiceName:
		return EMPTY
	case condition.Key_AwsPrincipalServiceNamesList:
		// TODO(nsiow) implement multivalue keys
		return EMPTY
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
	// we only model IAM entities; never services
	case
		condition.Key_AwsRequestCalledVia,
		condition.Key_AwsRequestCalledViaFirst,
		condition.Key_AwsRequestCalledViaLast,
		condition.Key_AwsRequestViaAwsService,
		condition.Key_AwsRequestSourceArn,
		condition.Key_AwsRequestSourceAccount,
		condition.Key_AwsRequestSourceOrgPaths,
		condition.Key_AwsRequestSourceOrgId:
		break
	case condition.Key_AwsRequestCurrentTime:
		return time.Now().UTC().Format(TIME_FORMAT)
	case condition.Key_AwsRequestEpochTime:
		// TODO(nsiow) make sure we are not losing accuracy
		epoch := int(time.Now().Unix())
		return strconv.Itoa(epoch)
	case condition.Key_AwsRequestReferer:
		break
	case condition.Key_AwsRequestRequestedRegion:
		break
	case condition.Key_AwsRequestTagKeys:
		// FIXME(nsiow) implement multi value key retrieval
		// maybe this whole function should just return []string?
		break
	case condition.Key_AwsRequestSecureTransport:
		break
	case condition.Key_AwsRequestUserAgent:
		break
	}

	return ac.Properties[key]
}

// Resolve resolves and replaces all IAM variables within the provided values
func (ac *AuthContext) Resolve(value string) string {
	matches := VariableExpansionRegex.FindAllStringSubmatch(value, -1)
	for _, match := range matches {
		if len(match) != 2 {
			panic(fmt.Sprintf("variable substitution choked on input: %s", value))
		}

		placeholder := match[0]
		variable := match[1]
		resolved := ac.Key(variable)
		value = strings.ReplaceAll(value, placeholder, resolved)
	}

	return value
}

// extractMultiValue defines how to create multivalue strings from single value ones
func (ac *AuthContext) extractMultiValue(v string) []string {
	return strings.Split(v, ",")
}

// extractTag defines how to get the value of the requested tag
func (ac *AuthContext) extractTag(key string, tags []entities.Tag) string {
	// Determine tag key
	components := strings.Split(key, "/")
	if len(components) != 2 {
		return ""
	}
	tagKey := components[1]

	for _, tag := range tags {
		if tag.Key == tagKey {
			return tag.Value
		}
	}

	return EMPTY
}
