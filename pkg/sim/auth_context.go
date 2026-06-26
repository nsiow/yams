package sim

import (
	"fmt"
	"regexp"
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
var VariableExpansionRegex = regexp.MustCompile(`\${([^}]+)}`)

// ConditionKey retrieves the value for the requested key from the AuthContext.
// Key lookups are case-insensitive for the key name portion; tag keys (after '/') remain
// case-sensitive per AWS behavior.
func (ac *AuthContext) ConditionKey(key string, opts Options) string {
	value, _ := ac.conditionKey(key, opts)
	return value
}

func (ac *AuthContext) HasConditionKey(key string, opts Options) bool {
	_, ok := ac.conditionKey(key, opts)
	return ok
}

func (ac *AuthContext) HasAnyKey(key string, opts Options) bool {
	if ac.HasConditionKey(key, opts) {
		return true
	}
	return ac.HasMultiKey(key, opts)
}

func (ac *AuthContext) conditionKey(key string, opts Options) (string, bool) {

	// ---------------------------------------------------------------------------------------------
	// Allow manual overrides
	// ---------------------------------------------------------------------------------------------

	value, ok := ac.Properties.Check(key)
	if ok && ac.supportsKey(key) {
		return value, true
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
		value, ok := ac.Properties.Check(key)
		return value, ok

		// ---------------------------------------------------------------------------------------------
		// Global keys; special handling
	// ---------------------------------------------------------------------------------------------

	case condkey.PrincipalArn:
		if ac.Principal == nil {
			return EMPTY, false
		}
		return ac.Principal.Arn, true
	case condkey.PrincipalAccount:
		if ac.Principal == nil {
			return EMPTY, false
		}
		return ac.Principal.AccountId, true
	case condkey.PrincipalIsAwsService:
		return "false", true // we do not support simulation for AWS services
	case condkey.PrincipalServiceName:
		return EMPTY, false // we do not support simulation for AWS services
	case condkey.PrincipalType:
		if ac.Principal == nil {
			return EMPTY, false
		}
		value := ac.principalType()
		return value, value != EMPTY
	case condkey.ResourceAccount:
		if ac.Resource == nil {
			return EMPTY, false
		}
		return ac.Resource.AccountId, true
	case condkey.CurrentTime:
		return ac.now().UTC().Format(DEFAULT_TIME_FORMAT), true
	case condkey.EpochTime:
		return strconv.FormatInt(ac.now().Unix(), 10), true
	case condkey.PrincipalOrgId:
		if ac.Principal == nil {
			return EMPTY, false
		}
		return ac.Principal.Account.OrgId, ac.Principal.Account.OrgId != EMPTY
	case condkey.ResourceOrgId:
		if ac.Resource == nil {
			return EMPTY, false
		}
		return ac.Resource.Account.OrgId, ac.Resource.Account.OrgId != EMPTY

	// ---------------------------------------------------------------------------------------------
	// Global key prefixes; special handling
	// ---------------------------------------------------------------------------------------------

	case condkey.PrincipalTagPrefix:
		if ac.Principal == nil {
			return EMPTY, false
		}
		return ac.extractTagValue(key, ac.Principal.Tags)
	}

	// ---------------------------------------------------------------------------------------------
	// SAR check
	// ---------------------------------------------------------------------------------------------

	// If it's not a global condition key, then we need to check the authorization reference
	if !opts.SkipServiceAuthorizationValidation && !ac.supportsKey(normalizedKey) {
		return EMPTY, false
	}

	// ---------------------------------------------------------------------------------------------
	// Local keys; prefix handling
	// ---------------------------------------------------------------------------------------------

	switch normalizedPrefix {
	case condkey.RequestTagPrefix:
		value, ok := ac.Properties.Check(key)
		return value, ok
	case condkey.ResourceTagPrefix:
		if ac.Resource == nil {
			return EMPTY, false
		}
		return ac.extractTagValue(key, ac.Resource.Tags)
	}

	if strings.HasSuffix(normalizedPrefix, ":resourceaccount") && ac.Resource != nil {
		return ac.Resource.AccountId, true
	}

	if isServiceResourceTagKey(normalizedPrefix) {
		if ac.Resource == nil {
			return EMPTY, false
		}
		return ac.extractTagValue(key, ac.Resource.Tags)
	}

	if isServiceRequestTagKey(normalizedPrefix) {
		if value, ok := ac.Properties.Check(key); ok {
			return value, true
		}
		if strings.Contains(key, "/") {
			_, tagKey, _ := strings.Cut(key, "/")
			value, ok := ac.Properties.Check(condkey.RequestTagPrefix + "/" + tagKey)
			return value, ok
		}
		return EMPTY, false
	}

	// ---------------------------------------------------------------------------------------------
	// Local keys; default handling
	// ---------------------------------------------------------------------------------------------

	value, ok = ac.Properties.Check(key)
	return value, ok
}

// MultiKey retrieves the values for the requested key from the AuthContext
func (ac *AuthContext) MultiKey(key string, opts Options) []string {
	values, _ := ac.multiKey(key, opts)
	return values
}

func (ac *AuthContext) HasMultiKey(key string, opts Options) bool {
	_, ok := ac.multiKey(key, opts)
	return ok
}

func (ac *AuthContext) multiKey(key string, opts Options) ([]string, bool) {

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
		if ac.Principal == nil {
			return nil, false
		}
		return ac.Principal.Account.OrgPaths, len(ac.Principal.Account.OrgPaths) > 0
	case condkey.ResourceOrgPaths:
		if ac.Resource == nil {
			return nil, false
		}
		return ac.Resource.Account.OrgPaths, len(ac.Resource.Account.OrgPaths) > 0
	}

	// ---------------------------------------------------------------------------------------------
	// SAR check
	// ---------------------------------------------------------------------------------------------

	if !opts.SkipServiceAuthorizationValidation && !ac.supportsKey(normalizedKey) {
		return nil, false
	}

	// ---------------------------------------------------------------------------------------------
	// Local keys; default handling
	// ---------------------------------------------------------------------------------------------

	values, ok := ac.MultiValueProperties.Check(key)
	return values, ok
}

// Substitute resolves and replaces all IAM variables within the provided values
func (ac *AuthContext) Substitute(value string, opts Options) string {
	return ac.SubstituteWithVersion(value, opts, "2012-10-17")
}

func (ac *AuthContext) SubstituteWithVersion(value string, opts Options, version string) string {
	if version != "" && version != "2012-10-17" {
		return value
	}

	matches := VariableExpansionRegex.FindAllStringSubmatch(value, -1)
	for _, match := range matches {

		placeholder := match[0]
		variable, fallback, hasFallback := parsePolicyVariable(match[1])
		resolved, ok := ac.conditionKey(variable, opts)
		if !ok {
			if !hasFallback {
				continue
			}
			resolved = fallback
		}
		value = strings.ReplaceAll(value, placeholder, resolved)
	}

	return value
}

func parsePolicyVariable(expr string) (string, string, bool) {
	switch expr {
	case "*", "?", "$":
		return expr, expr, true
	}

	variable, fallback, ok := strings.Cut(expr, ",")
	if !ok {
		return strings.TrimSpace(expr), "", false
	}

	fallback = strings.TrimSpace(fallback)
	fallback = strings.Trim(fallback, `"'`)
	return strings.TrimSpace(variable), fallback, true
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
	if i := strings.IndexByte(key, '/'); i >= 0 {
		return strings.ToLower(key[:i]) + "/" + key[i+1:]
	}
	return strings.ToLower(key)
}

// keyPrefix returns the prefix portion of the condition key, sans any attribute-getters
// afterwards; e.g. aws:RequestTag/foo becomes aws:RequestTag
func keyPrefix(key string) string {
	if i := strings.IndexByte(key, '/'); i >= 0 {
		return key[:i]
	}
	return key
}

func isServiceResourceTagKey(key string) bool {
	return strings.HasSuffix(key, ":resourcetag") ||
		key == "s3:existingobjecttag" ||
		key == "s3:buckettag"
}

func isServiceRequestTagKey(key string) bool {
	return strings.HasSuffix(key, ":requesttag") ||
		key == "s3:requestobjecttag"
}

// extractTag defines how to get the value of the requested tag
// TODO(nsiow) figure out if slashes are allowed in tag keys
func (ac *AuthContext) extractTag(key string, tags []entities.Tag) string {
	value, _ := ac.extractTagValue(key, tags)
	return value
}

func (ac *AuthContext) extractTagValue(key string, tags []entities.Tag) (string, bool) {
	i := strings.IndexByte(key, '/')
	if i < 0 || i == len(key)-1 {
		return EMPTY, false
	}
	tagKey := key[i+1:]

	for _, tag := range tags {
		if tag.Key == tagKey {
			return tag.Value, true
		}
	}

	return EMPTY, false
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
func (ac *AuthContext) supportsKey(key string) bool {
	normalizedKey := normalizeKey(key)
	normalizedPrefix := keyPrefix(normalizedKey)

	// First, check for global condition keys
	if condkey.IsGlobalConditionKey(normalizedPrefix) {
		return true
	}

	// Second, check if action supports key directly
	if ac.Action == nil {
		return true
	}
	for _, conditionKey := range ac.Action.ActionConditionKeys {
		if conditionKeyMatches(normalizedKey, normalizedPrefix, conditionKey) {
			return true
		}
	}

	// Otherwise check for each matched resource
	for _, resource := range ac.Action.Resources {
		for _, format := range resource.ARNFormats {
			if ac.Resource == nil {
				continue
			}
			// Use pre-split segments if available
			var match bool
			if len(ac.Resource.ArnSegments) > 0 {
				match = wildcard.MatchSegmentsPreSplit(format, ac.Resource.ArnSegments)
			} else {
				match = wildcard.MatchSegments(format, ac.Resource.Arn)
			}
			if match {
				for _, conditionKey := range resource.ConditionKeys {
					if conditionKeyMatches(normalizedKey, normalizedPrefix, conditionKey) {
						return true
					}
				}
			}
		}
	}

	return false
}

func conditionKeyMatches(key, prefix, supported string) bool {
	supported = normalizeKey(supported)
	if key == supported || prefix == supported {
		return true
	}
	if keyPrefix(supported) != prefix {
		return false
	}
	return strings.Contains(supported, "<key>") ||
		strings.Contains(supported, "${tagkey}")
}
