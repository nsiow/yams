package sim

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/nsiow/yams/pkg/aws/sar/types"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/loaders/awsconfig"
	condkey "github.com/nsiow/yams/pkg/policy/condition/keys"
	"github.com/nsiow/yams/pkg/sim/wildcard"
)

// TODO(nsiow) decide if principal/resource should be pointers or values; if pointers, implement
//             sufficient null checks

// AuthContext defines the tertiary context of a request that can be used for authz decisions
// TODO(nsiow) decide if this should be public or private type
type AuthContext struct {
	Action    *types.Action
	Principal *entities.FrozenPrincipal
	Resource  *entities.FrozenResource

	Time                 time.Time
	Properties           Bag[string]
	MultiValueProperties Bag[[]string]
}

// Static values
const (
	DEFAULT_TIME_FORMAT = "2006-01-02T15:04:05"
	EMPTY               = ""
	TRUE                = "true"
	FALSE               = "false"
)

var TIME_FORMATS = []string{
	"2006",
	"2006-01",
	"2006-01-02",
	"2006-01-02T15:04",
	"2006-01-02T15:04-0700",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05-0700",
	"2006-01-02T15:04:05.999",
	"2006-01-02T15:04:05.999Z",
	"2006-01-02T15:04:05.999-0700",
}

// VariableExpansionRegex defines the variable to use for expanding policy variables
var VariableExpansionRegex = regexp.MustCompile(`\${([a-zA-Z0-9]+:\S+?)}`)

// ConditionKey retrieves the value for the requested key from the AuthContext
// TODO(nsiow) key retrieval should be case insensitive... I think
// TODO(nsiow) support Trace object here for even lower level debugging
func (ac *AuthContext) ConditionKey(key string, opts Options) string {

	// ---------------------------------------------------------------------------------------------
	// Allow manual overrides
	// ---------------------------------------------------------------------------------------------

	value, ok := ac.Properties.Check(key)
	if ok && ac.supportsKey(key) {
		return value
	}

	// ---------------------------------------------------------------------------------------------
	// Normalize inputs
	// ---------------------------------------------------------------------------------------------

	normalizedKey := normalizeKey(key)
	normalizedPrefix := keyPrefix(normalizedKey)

	switch normalizedPrefix {

	// ---------------------------------------------------------------------------------------------
	// Global keys; default handling
	// ---------------------------------------------------------------------------------------------

	case
		condkey.CalledViaFirst,
		condkey.CalledViaLast,
		condkey.Ec2InstanceSourcePrivateIPv4,
		condkey.Ec2InstanceSourceVpc,
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
		// TODO(nsiow) implement userid/username where possible for principals
		condkey.UserId,
		condkey.Username,
		condkey.ViaAwsService,
		condkey.VpcSourceIp:
		return ac.Properties.Get(key)

	// ---------------------------------------------------------------------------------------------
	// Global keys; special handling
	// ---------------------------------------------------------------------------------------------

	case condkey.PrincipalArn:
		return ac.Principal.Arn
	case condkey.PrincipalAccount:
		return ac.Principal.AccountId
	case condkey.PrincipalIsAwsService:
		return "false" // we do not support simulation for AWS services
	case condkey.PrincipalServiceName:
		return EMPTY // we do not support simulation for AWS services
	case condkey.PrincipalType:
		return ac.principalType()
	case condkey.ResourceAccount:
		return ac.Resource.AccountId
	case condkey.CurrentTime:
		return ac.now().UTC().Format(DEFAULT_TIME_FORMAT)
	case condkey.EpochTime:
		return strconv.FormatInt(ac.now().Unix(), 10)
	case condkey.PrincipalOrgId:
		return ac.Principal.Account.OrgId
	case condkey.ResourceOrgId:
		return ac.Resource.Account.OrgId

	// ---------------------------------------------------------------------------------------------
	// Global key prefixes; special handling
	// ---------------------------------------------------------------------------------------------

	case condkey.PrincipalTagPrefix:
		return ac.extractTag(key, ac.Principal.Tags)
	}

	// ---------------------------------------------------------------------------------------------
	// SAR check
	// ---------------------------------------------------------------------------------------------

	// If it's not a global condition key, then we need to check the authorization reference
	if !opts.SkipServiceAuthorizationValidation && !ac.supportsKey(normalizedPrefix) {
		return EMPTY
	}

	// ---------------------------------------------------------------------------------------------
	// Local keys; prefix handling
	// ---------------------------------------------------------------------------------------------

	switch normalizedPrefix {
	case condkey.RequestTagPrefix:
		return ac.Properties.Get(key)
	case condkey.ResourceTagPrefix:
		return ac.extractTag(key, ac.Resource.Tags)
	}

	// ---------------------------------------------------------------------------------------------
	// Local keys; default handling
	// ---------------------------------------------------------------------------------------------

	return ac.Properties.Get(key)
}

// MultiKey retrieves the values for the requested key from the AuthContext
func (ac *AuthContext) MultiKey(key string, opts Options) []string {

	normalizedKey := normalizeKey(key)
	normalizedPrefix := keyPrefix(normalizedKey)

	// ---------------------------------------------------------------------------------------------
	// Global keys; default handling
	// ---------------------------------------------------------------------------------------------

	switch normalizedPrefix {
	case condkey.PrincipalServiceNamesList,
		condkey.CalledVia,
		condkey.TagKeys,
		condkey.SourceOrgPaths:
		break
	case condkey.PrincipalOrgPaths:
		return ac.Principal.Account.OrgPaths
	case condkey.ResourceOrgPaths:
		return ac.Resource.Account.OrgPaths
	}

	// ---------------------------------------------------------------------------------------------
	// SAR check
	// ---------------------------------------------------------------------------------------------

	if !opts.SkipServiceAuthorizationValidation && !ac.supportsKey(normalizedPrefix) {
		return nil
	}

	// ---------------------------------------------------------------------------------------------
	// Local keys; default handling
	// ---------------------------------------------------------------------------------------------

	return ac.MultiValueProperties.Get(key)
}

// Substitute resolves and replaces all IAM variables within the provided values
func (ac *AuthContext) Substitute(value string, opts Options) string {
	matches := VariableExpansionRegex.FindAllStringSubmatch(value, -1)
	for _, match := range matches {

		placeholder := match[0]
		variable := match[1]
		resolved := ac.ConditionKey(variable, opts)
		value = strings.ReplaceAll(value, placeholder, resolved)
	}

	return value
}

// Validate checks that the given AuthContext is valid and ready for simulation
func (ac *AuthContext) Validate(opts Options) error {
	// Handle the case where no principal is provided
	if ac.Principal == nil {
		return fmt.Errorf("AuthContext is missing Principal")
	}

	// Handle the case where no action is provided
	if ac.Action == nil {
		return fmt.Errorf("AuthContext is missing Action")
	}

	// All the remainder of the checks are SAR validations; skip if we disabled them
	if opts.SkipServiceAuthorizationValidation {
		return nil
	}

	// Handle the case where a resource is provided for a resource-less call
	if !ac.Action.HasTargets() && ac.Resource != nil {
		return fmt.Errorf("API call %s accepts no resources but was provided: %v",
			ac.Action.ShortName(), *ac.Resource)
	}

	// Handle the case where a call requires a resource but none is provided
	if ac.Action.HasTargets() && ac.Resource == nil {
		return fmt.Errorf("API call %s requires resources but none were provided",
			ac.Action.ShortName())
	}

	// Handle the case where the wrong resources are provided for the particular call
	if ac.Action.HasTargets() && !ac.Action.Targets(ac.Resource.Arn) {
		return fmt.Errorf(
			"resource ARN '%s' does not match any of allowed patterns for API call '%s': %v",
			ac.Resource.Arn, ac.Action.ShortName(), ac.Action.Resources)
	}

	return nil
}

// now returns the auth context's current frame of reference for the current time
func (ac *AuthContext) now() time.Time {
	// TODO(nsiow) wrap in DoOnce?
	if ac.Time.IsZero() {
		ac.Time = time.Now()
	}

	return ac.Time
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

// keyPrefix returns the prefix portion of the condition key, sans any attribute-getters
// afterwards; e.g. aws:RequestTag/foo becomes aws:RequestTag
func keyPrefix(key string) string {
	substr := strings.SplitN(key, "/", 2)
	return substr[0]
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
		return "AssumedRole"
	case awsconfig.CONST_TYPE_AWS_IAM_USER:
		return "User"
	default:
		return EMPTY
	}
}

// supportsKey consults the SAR package to determine whether or not the requested key is supported
// for the simulated API call
// TODO(nsiow) perform condition key type validation
func (ac *AuthContext) supportsKey(key string) bool {
	normalizedPrefix := keyPrefix(key)

	// First, check for global condition keys
	if condkey.IsGlobalConditionKey(normalizedPrefix) {
		return true
	}

	// Second, check if action supports key directly
	if ac.Action == nil || slices.Contains(ac.Action.ActionConditionKeys, normalizedPrefix) {
		return true
	}

	// Otherwise check for each matched resource
	for _, resource := range ac.Action.Resources {
		for _, format := range resource.ARNFormats {
			if ac.Resource != nil && wildcard.MatchSegments(format, ac.Resource.Arn) {
				if slices.Contains(resource.ConditionKeys, normalizedPrefix) {
					return true
				}
			}
		}
	}

	return false
}
