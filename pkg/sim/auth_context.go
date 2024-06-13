package sim

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/loaders/awsconfig"
	"github.com/nsiow/yams/pkg/policy/condition"
)

// TODO(nsiow) decide if principal/resource should be pointers or values; if pointers, implement null checks
// TODO(nsiow) check condition key reference for single vs multi values

// AuthContext defines the tertiary context of a request that can be used for authz decisions
type AuthContext struct {
	Time      time.Time
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
var VariableExpansionRegex = regexp.MustCompile(`\${([a-zA-Z0-9]+:\S+?)}`)

func (ac *AuthContext) Key(key string) string {
	// Try handling prefixes first...
	switch {
	case strings.HasPrefix(key, condition.Key_AwsPrincipalTagPrefix):
		return ac.extractTag(key, ac.Principal.Tags)
	case strings.HasPrefix(key, condition.Key_AwsResourceTagPrefix):
		return ac.extractTag(key, ac.Resource.Tags)
	case strings.HasPrefix(key, condition.Key_AwsRequestTagPrefix):
		break // it's not a prefix, so process it as a key
	}

	// ... otherwise handle as a static key
	switch key {
	case condition.Key_AwsPrincipalArn:
		return ac.Principal.Arn
	case condition.Key_AwsPrincipalAccount:
		return ac.Principal.Account
	case condition.Key_AwsPrincipalIsAwsService:
		break
	case condition.Key_AwsPrincipalServiceName:
		break
	case condition.Key_AwsPrincipalType:
		return ac.principalType()
	case condition.Key_AwsResourceAccount:
		return ac.Resource.Account
	case condition.Key_AwsRequestCurrentTime:
		return ac.now().UTC().Format(TIME_FORMAT)
	case condition.Key_AwsRequestEpochTime:
		// TODO(nsiow) make sure we are not losing accuracy
		epoch := int(ac.now().Unix())
		return strconv.Itoa(epoch)

	// FIXME(nsiow) implement multi value key retrieval
	// case condition.Key_AwsRequestTagKeys:
	// 	break
	// case condition.Key_AwsPrincipalServiceNamesList:
	//  break

	// TODO(nsiow) revisit when we have org support
	// case condition.Key_AwsPrincipalOrgPaths:
	// 	break
	// case condition.Key_AwsPrincipalOrgId:
	// 	break
	// case condition.Key_AwsResourceOrgPaths:
	// 	break
	// case condition.Key_AwsResourceOrgId:
	// 	break

	// We'll enumerate these for potential special handling in the future, but otherwise just use
	// default behavior
	case
		condition.Key_AwsNetworkSourceIp,
		condition.Key_AwsNetworkSourceVpc,
		condition.Key_AwsNetworkSourceVpce,
		condition.Key_AwsNetworkVpcSourceIp,
		condition.Key_AwsPrincipalServiceNamesList,
		condition.Key_AwsPrincipalUserId,
		condition.Key_AwsPrincipalUsername,
		condition.Key_AwsRequestCalledVia,
		condition.Key_AwsRequestCalledViaFirst,
		condition.Key_AwsRequestCalledViaLast,
		condition.Key_AwsRequestReferer,
		condition.Key_AwsRequestRequestedRegion,
		condition.Key_AwsRequestSecureTransport,
		condition.Key_AwsRequestSourceAccount,
		condition.Key_AwsRequestSourceArn,
		condition.Key_AwsRequestSourceOrgId,
		condition.Key_AwsRequestSourceOrgPaths,
		condition.Key_AwsRequestUserAgent,
		condition.Key_AwsRequestViaAwsService,
		condition.Key_AwsSessionFederatedProvider,
		condition.Key_AwsSessionMfaAge,
		condition.Key_AwsSessionMfaPresent,
		condition.Key_AwsSessionRoleDelivery,
		condition.Key_AwsSessionSourceIdentity,
		condition.Key_AwsSessionSourceInstanceArn,
		condition.Key_AwsSessionSourceIpv4,
		condition.Key_AwsSessionSourceVpc,
		condition.Key_AwsSessionTokenIssueTime:
		break
	}

	return ac.Properties[key]
}

// Resolve resolves and replaces all IAM variables within the provided values
func (ac *AuthContext) Resolve(value string) string {
	matches := VariableExpansionRegex.FindAllStringSubmatch(value, -1)
	for _, match := range matches {

		placeholder := match[0]
		variable := match[1]
		resolved := ac.Key(variable)
		fmt.Printf("placeholder = %s, variable = %s, resolved = %s\n", placeholder, variable, resolved)
		value = strings.ReplaceAll(value, placeholder, resolved)
	}

	return value
}

// TODO(nsiow) implement multivalue keys
// extractMultiValue defines how to create multivalue strings from single value ones
// func (ac *AuthContext) extractMultiValue(v string) []string {
//   return strings.Split(v, ",")
// }

// extractTag defines how to get the value of the requested tag
// TODO(nsiow) figure out if slashes are allowed in tag keys
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

// principalType determines the type of the Principal for use with the aws:PrincipalType key
func (ac *AuthContext) principalType() string {
	switch ac.Principal.Type {
	case awsconfig.CONST_TYPE_AWS_IAM_ROLE:
		return "Role"
	case awsconfig.CONST_TYPE_AWS_IAM_USER:
		return "User"
	default:
		return EMPTY
	}
}

// now returns the auth context's current frame of reference for the current time
func (ac *AuthContext) now() time.Time {
	// TODO(nsiow) wrap in DoOnce?
	if ac.Time.IsZero() {
		ac.Time = time.Now()
	}

	return ac.Time
}
