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
	Properties           *PropertyBag[string]
	MultiValueProperties *PropertyBag[[]string]
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
// TODO(nsiow) support Trace object here for even lower level debugging
func (ac *AuthContext) Key(key string) string {
	key = normalizeKey(key)

	// // TODO(nsiow) implement some sort of option around this for faster, less-strict sims
	// if !ac.SupportsKey(key) {
	// 	return EMPTY
	// }

	// Try handling prefixes first...
	switch {
	case strings.HasPrefix(key, condkey.PrincipalTagPrefix):
		return ac.extractTag(key, ac.Principal.Tags)
	case strings.HasPrefix(key, condkey.ResourceTagPrefix):
		return ac.extractTag(key, ac.Resource.Tags)
		// For RequestTags/, we will process like a standard key
	}

	// TODO(nsiow) handle case where ${aws:PrincipalTag/foo} is in Resource= field, etc

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
		return ac.Resource.AccountId
	case condkey.CurrentTime:
		return ac.now().UTC().Format(TIME_FORMAT)
	case condkey.EpochTime:
		return strconv.FormatInt(ac.now().Unix(), 10)
	case condkey.PrincipalOrgId:
		return ac.Principal.Account.OrgId
	case condkey.ResourceOrgId:
		return ac.Resource.Account.OrgId

	// We'll enumerate these for potential special handling in the future, but otherwise just use
	// default behavior
	case
		condkey.CalledViaFirst,
		condkey.CalledViaLast,
		condkey.Ec2InstanceSourcePrivateIPv4,
		condkey.FederatedProvider,
		condkey.MultiFactorAuthAge,
		condkey.MultiFactorAuthPresent,
		condkey.PrincipalServiceNamesList,
		condkey.Referer,
		condkey.RequestedRegion,
		condkey.RoleDelivery,
		condkey.SecureTransport,
		condkey.SourceAccount,
		condkey.SourceArn,
		condkey.SourceIdentity,
		condkey.SourceInstanceArn,
		condkey.SourceIp,
		condkey.SourceOrgId,
		condkey.SourceVpc,
		condkey.SourceVpce,
		condkey.TokenIssueTime,
		condkey.UserAgent,
		condkey.UserId,
		condkey.Username,
		condkey.ViaAwsService,
		condkey.VpcSourceIp:
		break
	}

	if ac.Properties == nil {
		return EMPTY
	}
	return ac.Properties.Get(key)
}

// normalizeKey performs any required key normalization to process the provided key
func normalizeKey(key string) string {
	// TODO(nsiow) this is a rough approximation
	substr := strings.SplitN(key, "/", 2)
	switch len(substr) {
	case 1:
		return strings.ToLower(key)
	default:
		return strings.ToLower(substr[0]) + "/" + substr[1]
	}
}

// MultiKey retrieves the values for the requested key from the AuthContext
func (ac *AuthContext) MultiKey(key string) []string {
	key = normalizeKey(key)

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

	if ac.MultiValueProperties == nil {
		return nil
	}
	return ac.MultiValueProperties.Get(key)
}

// SupportsKey consults the SAR package to determine whether or not the requested key is even
// supported for the simulated API call
func (ac *AuthContext) SupportsKey(key string) bool {
	return true
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
	components := strings.SplitN(key, "/", 2)
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
