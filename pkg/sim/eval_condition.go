package sim

import (
	"errors"
	"fmt"
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

// Error indicating that an unknown Condition operator was specified
var ErrUnknownOperator = errors.New("unknown operation")

// Error indicating that an unknown Condition key was specified
var ErrUnknownConditionKey = errors.New("unknown condition key")

// Compare defines a function used to compare a value to a single other value
//
// The function should take in two strings where `left` is the observed value and `right` is what
// we are trying to match against
type Compare = func(trc *Trace, left string, right string) bool

// TODO(nsiow) revisit the naming of the constructs below

// ConditionOperator defines a function used to compare a value against a list of possible other
// values
type ConditionOperator = func(ac AuthContext, trc *Trace, left string, right policy.Value) bool

// ConditionEvaluator defines a function that can be used to fully evaluate a Condition block
type ConditionEvaluator = func(ac AuthContext, trc *Trace, key string, right policy.Value) bool

// ConditionPredicate defines the predicate that must be true in order for a condition to succeed
type ConditionPredicate = func(ConditionOperator) ConditionEvaluator

// ConditionMod defines a function which wraps a ConditionOperator to change its behavior
type ConditionMod = func(ConditionOperator) ConditionOperator

// --------------------------------------------------------------------------------
// Mappings
// --------------------------------------------------------------------------------

// ConditionOperatorMap defines the mapping between operator names and functions
var ConditionOperatorMap = map[string]ConditionOperator{

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
}

// --------------------------------------------------------------------------------
// Condition evaluation functions
// --------------------------------------------------------------------------------

func Cond_MatchAny(f Compare) ConditionOperator {
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

// --------------------------------------------------------------------------------
// Condition modifiers
// --------------------------------------------------------------------------------

// Mod_Not inverts the provided ConditionOperator
func Mod_Not(f ConditionOperator) ConditionOperator {
	return func(ac AuthContext, trc *Trace, left string, right policy.Value) bool {
		return !f(ac, trc, left, right)
	}
}

// Mod_ResolveVariables resolves and replaces all IAM variables within the provided values
func Mod_ResolveVariables(f ConditionOperator) ConditionOperator {
	return func(ac AuthContext, trc *Trace, left string, right policy.Value) bool {
		for i := range right {
			right[i] = ac.Resolve(right[i])
		}
		return f(ac, trc, left, right)
	}
}

// Mod_MustExist defines a Condition modifier which returns false if the key is not found
func Mod_MustExist(f ConditionOperator) ConditionOperator {
	return func(ac AuthContext, trc *Trace, left string, right policy.Value) bool {
		if left == "" {
			return false
		}

		return f(ac, trc, left, right)
	}
}

// Mod_IfExists defines a Condition modifier which returns true if the key is not found
func Mod_IfExists(f ConditionOperator) ConditionOperator {
	return func(ac AuthContext, trc *Trace, left string, right policy.Value) bool {
		if left == "" {
			return true
		}

		return f(ac, trc, left, right)
	}
}

// Mod_ForAllValues defines a Condition modifier targeting match-all logic for multivalued
// conditions
func Mod_ForAllValues(f ConditionOperator) ConditionEvaluator {
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
func Mod_ForAnyValues(f ConditionOperator) ConditionEvaluator {
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
func Mod_ForSingleValue(f ConditionOperator) ConditionEvaluator {
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

// --------------------------------------------------------------------------------
// Externally facing functions
// --------------------------------------------------------------------------------

// ResolveConditionEvaluator takes in an operator name and resolves it to a function
//
// If the function could be resolved, the second return value is `true`. Otherwise, the second
// return value is `false`
func ResolveConditionEvaluator(op string) (ConditionEvaluator, bool) {
	// Determine the condition predicate
	var predicate ConditionPredicate
	if strings.HasPrefix(op, "ForAllValues:") {
		predicate = Mod_ForAllValues
		op = strings.TrimPrefix(op, "ForAllValues:")
	} else if strings.HasPrefix(op, "ForAnyValues:") {
		predicate = Mod_ForAnyValues
		op = strings.TrimPrefix(op, "ForAnyValues:")
	} else {
		predicate = Mod_ForSingleValue
	}

	// Handle function modifiers
	mods := []ConditionMod{}
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
	return predicate(f), true
}
