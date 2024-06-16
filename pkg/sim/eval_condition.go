package sim

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/policy/condition"
	"github.com/nsiow/yams/pkg/sim/wildcard"
)

// --------------------------------------------------------------------------------
// Setup
// --------------------------------------------------------------------------------

// TODO(nsiow) make better use of the two error types below

// Error indicating that an unknown Condition operator was specified
var ErrUnknownOperator = errors.New("unknown operation")

// Error indicating that an unknown Condition key was specified
var ErrUnknownConditionKey = errors.New("unknown condition key")

// Compare defines a function used to compare a value to a single other value
//
// The function should take in two strings where `left` is the observed value and `right` is what
// we are trying to match against
type Compare = func(trc *Trace, left string, right string) bool

// CondOuter defines a function that accepts a key name and set of values and evaluates the
// outcome of the condition
type CondOuter = func(ac AuthContext, trc *Trace, key string, right policy.Value) bool

// CondInner defines a function that accepts a left hand value and a right hand set of values
// and evaluates the outcome of the condition
type CondInner = func(ac AuthContext, trc *Trace, left string, right policy.Value) bool

// CondLift defines a function which "lifts" a ConditionInner operator
//
// This function effectively contains the logic to map the "key" parameter of a ConditionOuter
// function to the "left" parameter of a ConditionInner function
type CondLift = func(CondInner) CondOuter

// CondMod defines a function which wraps a ConditionOperator to change its behavior
type CondMod = func(CondInner) CondInner

// --------------------------------------------------------------------------------
// Mappings
// --------------------------------------------------------------------------------

// ConditionOperatorMap defines the mapping between operator names and functions
var ConditionOperatorMap = map[string]CondInner{

	// ------------------------------------------------------------------------------
	// String Functions
	// ------------------------------------------------------------------------------

	condition.StringEquals: Mod_ResolveVariables(
		Cond_MatchAny(
			Cond_StringEquals,
		),
	),
	condition.StringNotEquals: Mod_ResolveVariables(
		Mod_Not(
			Cond_MatchAny(
				Cond_StringEquals,
			),
		),
	),
	condition.StringEqualsIgnoreCase: Mod_ResolveVariables(
		Cond_MatchAny(
			Mod_IgnoreCase(
				Cond_StringEquals,
			),
		),
	),
	condition.StringNotEqualsIgnoreCase: Mod_ResolveVariables(
		Mod_Not(
			Cond_MatchAny(
				Mod_IgnoreCase(
					Cond_StringEquals,
				),
			),
		),
	),
	condition.StringLike: Mod_ResolveVariables(
		Cond_MatchAny(
			Cond_StringLike,
		),
	),
	condition.StringNotLike: Mod_ResolveVariables(
		Mod_Not(
			Cond_MatchAny(
				Cond_StringLike,
			),
		),
	),

	// ------------------------------------------------------------------------------
	// Numeric Functions
	// ------------------------------------------------------------------------------

	condition.NumericEquals: Cond_MatchAny(
		Mod_Number(
			Cond_NumericEquals,
		),
	),
	condition.NumericNotEquals: Mod_Not(
		Cond_MatchAny(
			Mod_Number(
				Cond_NumericEquals,
			),
		),
	),
	condition.NumericLessThan: Cond_MatchAny(
		Mod_Number(
			Cond_NumericLessThan,
		),
	),
	condition.NumericLessThanEquals: Cond_MatchAny(
		Mod_Number(
			Cond_NumericLessThanEquals,
		),
	),
	condition.NumericGreaterThan: Cond_MatchAny(
		Mod_Number(
			Cond_NumericGreaterThan,
		),
	),
	condition.NumericGreaterThanEquals: Cond_MatchAny(
		Mod_Number(
			Cond_NumericGreaterThanEquals,
		),
	),

	// ------------------------------------------------------------------------------
	// Date Functions
	// ------------------------------------------------------------------------------

	condition.DateEquals: Cond_MatchAny(
		Mod_Date(
			Cond_NumericEquals,
		),
	),
	condition.DateNotEquals: Mod_Not(
		Cond_MatchAny(
			Mod_Date(
				Cond_NumericEquals,
			),
		),
	),
	condition.DateLessThan: Cond_MatchAny(
		Mod_Date(
			Cond_NumericLessThan,
		),
	),
	condition.DateLessThanEquals: Cond_MatchAny(
		Mod_Date(
			Cond_NumericLessThanEquals,
		),
	),
	condition.DateGreaterThan: Cond_MatchAny(
		Mod_Date(
			Cond_NumericGreaterThan,
		),
	),
	condition.DateGreaterThanEquals: Cond_MatchAny(
		Mod_Date(
			Cond_NumericGreaterThanEquals,
		),
	),

	// ------------------------------------------------------------------------------
	// Bool Functions
	// ------------------------------------------------------------------------------

	condition.Bool: Mod_ResolveVariables(
		Cond_MatchAny(
			Mod_Bool(
				Mod_IgnoreCase(
					Cond_StringEquals,
				),
			),
		),
	),

	// ------------------------------------------------------------------------------
	// Binary Functions
	// ------------------------------------------------------------------------------

	condition.BinaryEquals: Cond_MatchAny(
		Mod_Binary(
			Cond_StringEquals,
		),
	),

	// ------------------------------------------------------------------------------
	// IP Address Functions
	// ------------------------------------------------------------------------------

	condition.IpAddress: Cond_MatchAny(
		Mod_Network(
			Cond_IpAddress,
		),
	),
	condition.NotIpAddress: Mod_Not(
		Cond_MatchAny(
			Mod_Network(
				Cond_IpAddress,
			),
		),
	),

	// ------------------------------------------------------------------------------
	// ARN Functions
	// ------------------------------------------------------------------------------

	condition.ArnEquals: Mod_ResolveVariables(
		Cond_MatchAny(
			Cond_ArnLike,
		),
	),
	condition.ArnNotEquals: Mod_ResolveVariables(
		Mod_Not(
			Cond_MatchAny(
				Cond_ArnLike,
			),
		),
	),
	condition.ArnLike: Mod_ResolveVariables(
		Cond_MatchAny(
			Cond_ArnLike,
		),
	),
	condition.ArnNotLike: Mod_ResolveVariables(
		Mod_Not(
			Cond_MatchAny(
				Cond_ArnLike,
			),
		),
	),
}

// --------------------------------------------------------------------------------
// Condition evaluation functions
// --------------------------------------------------------------------------------

func Cond_MatchAny(f Compare) CondInner {
	return func(_ AuthContext, trc *Trace, left string, right policy.Value) bool {
		for _, value := range right {
			if f(trc, left, value) {
				return true
			}
		}

		return false
	}
}

// --------------------------------------------------------------------------------
// Condition comparison functions
// --------------------------------------------------------------------------------

// Cond_StringEquals defines the `StringEquals` condition function
func Cond_StringEquals(trc *Trace, left, right string) bool {
	return left == right
}

// Cond_StringLike defines the `StringLike` condition function
func Cond_StringLike(trc *Trace, left, right string) bool {
	// TODO(nsiow) maybe swap ordering of arguments in matchWildcard to better match go stdlib
	return wildcard.MatchString(right, left)
}

// Cond_NumericEquals defines the `NumericEquals` condition function
func Cond_NumericEquals(trc *Trace, left, right int) bool {
	return left == right
}

// Cond_NumericLessThan defines the `NumericLessThan` condition function
func Cond_NumericLessThan(trc *Trace, left, right int) bool {
	return left < right
}

// Cond_NumericLessThanEquals defines the `NumericLessThanEquals` condition function
func Cond_NumericLessThanEquals(trc *Trace, left, right int) bool {
	return left <= right
}

// Cond_NumericGreaterThan defines the `NumericGreaterThan` condition function
func Cond_NumericGreaterThan(trc *Trace, left, right int) bool {
	return left > right
}

// Cond_NumericGreaterThanEquals defines the `NumericGreaterThanEquals` condition function
func Cond_NumericGreaterThanEquals(trc *Trace, left, right int) bool {
	return left >= right
}

// Cond_IpAddress defines the `IpAddress` condition function
func Cond_IpAddress(trc *Trace, left netip.Addr, right netip.Prefix) bool {
	return right.Contains(left)
}

// Cond_ArnLike defines the `ArnLike` condition function
func Cond_ArnLike(trc *Trace, left, right string) bool {
	return wildcard.MatchArn(right, left)
}

// --------------------------------------------------------------------------------
// Condition modifiers
// --------------------------------------------------------------------------------

// Mod_Not inverts the provided ConditionOperator
func Mod_Not(f CondInner) CondInner {
	return func(ac AuthContext, trc *Trace, left string, right policy.Value) bool {
		return !f(ac, trc, left, right)
	}
}

// Mod_ResolveVariables resolves and replaces all IAM variables within the provided values
func Mod_ResolveVariables(f CondInner) CondInner {
	return func(ac AuthContext, trc *Trace, left string, right policy.Value) bool {
		for i := range right {
			right[i] = ac.Resolve(right[i])
		}
		return f(ac, trc, left, right)
	}
}

// Mod_MustExist defines a Condition modifier which returns false if the key is not found
func Mod_MustExist(f CondInner) CondInner {
	return func(ac AuthContext, trc *Trace, left string, right policy.Value) bool {
		if left == "" {
			return false
		}

		return f(ac, trc, left, right)
	}
}

// Mod_IfExists defines a Condition modifier which returns true if the key is not found
func Mod_IfExists(f CondInner) CondInner {
	return func(ac AuthContext, trc *Trace, left string, right policy.Value) bool {
		if left == "" {
			return true
		}

		return f(ac, trc, left, right)
	}
}

// Mod_ForAllValues defines a Condition modifier targeting match-all logic for multivalued
// conditions
func Mod_ForAllValues(f CondInner) CondOuter {
	return func(ac AuthContext, trc *Trace, key string, right policy.Value) bool {
		lefts := ac.MultiKey(key)
		for _, left := range lefts {
			if !f(ac, trc, left, right) {
				return false
			}
		}

		return true
	}
}

// Mod_ForAnyValues defines a Condition modifier targeting match-any logic for multivalued
// conditions
func Mod_ForAnyValues(f CondInner) CondOuter {
	return func(ac AuthContext, trc *Trace, key string, right policy.Value) bool {
		lefts := ac.MultiKey(key)
		for _, left := range lefts {
			if f(ac, trc, left, right) {
				return true
			}
		}

		return false
	}
}

// Mod_ForSIngleValue defines a Condition modifier targeting match-any logic for single-valued
// conditions (the default)
func Mod_ForSingleValue(f CondInner) CondOuter {
	return func(ac AuthContext, trc *Trace, key string, right policy.Value) bool {
		left := ac.Key(key)
		return f(ac, trc, left, right)
	}
}

// Mod_IgnoreCase defines a Condition modifier which ignores character casing
func Mod_IgnoreCase(f Compare) Compare {
	return func(trc *Trace, left, right string) bool {
		return f(trc, strings.ToLower(left), strings.ToLower(right))
	}
}

// Mod_Number converts the string inputs to numbers, allowing numerical comparisons
func Mod_Number(f func(*Trace, int, int) bool) Compare {
	return func(trc *Trace, left, right string) bool {
		nLeft, err := strconv.Atoi(left)
		if err != nil {
			// TODO(nsiow) find a good place to log errors
			return false
		}

		nRight, err := strconv.Atoi(right)
		if err != nil {
			// TODO(nsiow) find a good place to log errors
			return false
		}

		return f(trc, nLeft, nRight)
	}
}

// parseEpochFromString is a helper function allowing us to extract an epoch timestamp from a
// string
func parseEpochFromString(s string) (int, error) {
	asDatetime, err := time.Parse(TIME_FORMAT, s)
	if err == nil {
		return int(asDatetime.Unix()), nil
	}

	asEpoch, err := strconv.Atoi(s)
	if err == nil {
		return asEpoch, nil
	}

	return -1, fmt.Errorf("unable to parse time '%s' as either datetime or epoch", s)
}

// Mod_Date converts the string inputs to dates, allowing datewise comparisons
func Mod_Date(f func(*Trace, int, int) bool) Compare {
	return func(trc *Trace, left, right string) bool {
		nLeft, err := parseEpochFromString(left)
		if err != nil {
			// TODO(nsiow) find a good place to log errors
			return false
		}

		nRight, err := parseEpochFromString(right)
		if err != nil {
			// TODO(nsiow) find a good place to log errors
			return false
		}

		return f(trc, nLeft, nRight)
	}
}

// Mod_Bool converts the string inputs to bools, allowing boolean operations
func Mod_Bool(f func(*Trace, string, string) bool) Compare {
	return func(trc *Trace, left, right string) bool {
		bLeft := strings.ToLower(left)
		if bLeft != TRUE && bLeft != FALSE {
			return false
		}

		bRight := strings.ToLower(right)
		if bRight != TRUE && bRight != FALSE {
			return false
		}

		return f(trc, bLeft, bRight)
	}
}

// Mod_Binary validates and forwards on the base64 encoded values, allowing binary expressions
//
// We reuse the string operators for this rather than a byte-by-byte comparison for ease, but for
// slightly faster comparison we should perform the byte-by-byte comparison to avoid the base64
// encoding overhead
func Mod_Binary(f func(*Trace, string, string) bool) Compare {
	return func(trc *Trace, left, right string) bool {
		_, err := base64.StdEncoding.DecodeString(left)
		if err != nil {
			// TODO(nsiow) add to Trace
			return false
		}

		_, err = base64.StdEncoding.DecodeString(right)
		if err != nil {
			// TODO(nsiow) add to Trace
			return false
		}

		return f(trc, left, right)
	}
}

// Mod_Network converts the incoming strings into IP addresses/nets, allowing network expressions
func Mod_Network(f func(*Trace, netip.Addr, netip.Prefix) bool) Compare {
	return func(trc *Trace, left, right string) bool {
		addr, err := netip.ParseAddr(left)
		if err != nil {
			// TODO(nsiow) add to Trace
			return false
		}

		network, err := netip.ParsePrefix(right)
		if err != nil {
			// TODO(nsiow) add to Trace
			return false
		}

		return f(trc, addr, network)
	}
}

// --------------------------------------------------------------------------------
// Externally facing functions
// --------------------------------------------------------------------------------

// ResolveConditionEvaluator takes in an operator name and resolves it to a function
//
// If the function could be resolved, the second return value is `true`. Otherwise, the second
// return value is `false`
func ResolveConditionEvaluator(op string) (CondOuter, bool) {
	// Determine the condition lift
	var lift CondLift
	if strings.HasPrefix(op, "ForAllValues:") {
		lift = Mod_ForAllValues
		op = strings.TrimPrefix(op, "ForAllValues:")
	} else if strings.HasPrefix(op, "ForAnyValues:") {
		lift = Mod_ForAnyValues
		op = strings.TrimPrefix(op, "ForAnyValues:")
	} else {
		lift = Mod_ForSingleValue
	}

	// Handle function modifiers
	mods := []CondMod{}
	if strings.HasSuffix(op, "IfExists") && !strings.HasPrefix(op, "Null") {
		mods = append(mods, Mod_IfExists)
		op = strings.TrimSuffix(op, "IfExists")
	} else {
		mods = append(mods, Mod_MustExist)
	}

	// Attempt to look up function
	f, exists := ConditionOperatorMap[op]
	if !exists {
		return nil, false
	}

	// Apply modifiers
	for _, mod := range mods {
		f = mod(f)
	}
	return lift(f), true
}
