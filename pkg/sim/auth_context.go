package sim

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/loaders/awsconfig"
	condkey "github.com/nsiow/yams/pkg/policy/condition/keys"
)

// TODO(nsiow) decide if principal/resource should be pointers or values; if pointers, implement
//             sufficient null checks

// AuthContext defines the tertiary context of a request that can be used for authz decisions
type AuthContext struct {
	Time                 time.Time
	Action               string
	Principal            *entities.Principal
	Resource             *entities.Resource
	Properties           map[string]string
	MultiValueProperties map[string][]string
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

// Key retrieves the value for the requested key from the AuthContext
// TODO(nsiow) key retrieval should be case insensitive... I think
func (ac *AuthContext) Key(key string) string {
	// Try handling prefixes first...
	switch {
	case strings.HasPrefix(key, condkey.PrincipalTagPrefix):
		return ac.extractTag(key, ac.Principal.Tags)
	case strings.HasPrefix(key, condkey.ResourceTagPrefix):
		return ac.extractTag(key, ac.Resource.Tags)
		// For RequestTags/, we will process like a standard key
	}

	// ... otherwise handle as a static key
	switch key {
	case condkey.PrincipalArn:
		return ac.Principal.Arn
	case condkey.PrincipalAccount:
		return ac.Principal.AccountId
	case condkey.PrincipalIsAwsService:
		break
	case condkey.PrincipalServiceName:
		break
	case condkey.PrincipalType:
		return ac.principalType()
	case condkey.ResourceAccount:
		return ac.Resource.Account
	case condkey.CurrentTime:
		return ac.now().UTC().Format(TIME_FORMAT)
	case condkey.EpochTime:
		return strconv.FormatInt(ac.now().Unix(), 10)
	case condkey.PrincipalOrgId:
		return ac.Principal.Account.OrgId
	case condkey.ResourceOrgId:
		// FIXME(nsiow)
		panic("not yet implemented")

	// We'll enumerate these for potential special handling in the future, but otherwise just use
	// default behavior
	// TODO(nsiow) consider switching this to condkey.SourceIp etc, in a separate package
	case
		condkey.SourceIp,
		condkey.SourceVpc,
		condkey.SourceVpce,
		condkey.VpcSourceIp,
		condkey.PrincipalServiceNamesList,
		condkey.UserId,
		condkey.Username,
		condkey.CalledViaFirst,
		condkey.CalledViaLast,
		condkey.Referer,
		condkey.RequestedRegion,
		condkey.SecureTransport,
		condkey.SourceAccount,
		condkey.SourceArn,
		condkey.SourceOrgId,
		condkey.UserAgent,
		condkey.ViaAwsService,
		condkey.FederatedProvider,
		condkey.MultiFactorAuthAge,
		condkey.MultiFactorAuthPresent,
		condkey.RoleDelivery,
		condkey.SourceIdentity,
		condkey.SourceInstanceArn,
		condkey.Ec2InstanceSourcePrivateIPv4,
		condkey.TokenIssueTime:
		break
	}

	return ac.Properties[key]
}

// MultiKey retrieves the values for the requested key from the AuthContext
func (ac *AuthContext) MultiKey(key string) []string {
	switch key {
	case condkey.PrincipalServiceNamesList,
		condkey.CalledVia,
		condkey.TagKeys,
		condkey.SourceOrgPaths:
		break

		// TODO(nsiow) revisit when we have org support
		// case condkey.PrincipalOrgPaths:
		// 	break
		// case condkey.ResourceOrgPaths:
		// 	break
	}

	return ac.MultiValueProperties[key]
}

// Resolve resolves and replaces all IAM variables within the provided values
func (ac *AuthContext) Resolve(value string) string {
	matches := VariableExpansionRegex.FindAllStringSubmatch(value, -1)
	for _, match := range matches {

		placeholder := match[0]
		variable := match[1]
		resolved := ac.Key(variable)
		value = strings.ReplaceAll(value, placeholder, resolved)
	}

	return value
}

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
