package sim

import (
	"errors"
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

// ConditionOperator defines a function used to compare a value against a list of possible other
// values
type ConditionOperator = func(ac AuthContext, trc *Trace, left string, right policy.Value) bool

// ConditionMod defines a function which wraps a ConditionOperator
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
		Cond_MatchNone(
			Cond_StringEquals,
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
		Cond_MatchNone(
			Mod_IgnoreCase(
				Cond_StringEquals,
			),
		),
	),
	condition.StringLike: Mod_ResolveVariables(
		Cond_MatchAny(
			Cond_StringLike,
		),
	),
	condition.StringNotLike: Mod_ResolveVariables(
		Cond_MatchNone(
			Cond_StringLike,
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
	condition.NumericNotEquals: Cond_MatchNone(
		Mod_Number(
			Cond_NumericEquals,
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
	condition.DateNotEquals: Cond_MatchNone(
		Mod_Not(
			Mod_Date(
				Cond_NumericEquals,
			),
		),
	),
	condition.DateLessThan: Cond_MatchNone(
		Mod_Date(
			Cond_NumericLessThan,
		),
	),
	condition.DateLessThanEquals: Cond_MatchNone(
		Mod_Date(
			Cond_NumericLessThanEquals,
		),
	),
	condition.DateGreaterThan: Cond_MatchNone(
		Mod_Date(
			Cond_NumericGreaterThan,
		),
	),
	condition.DateGreaterThanEquals: Cond_MatchNone(
		Mod_Date(
			Cond_NumericGreaterThanEquals,
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

func Cond_MatchNone(f Compare) ConditionOperator {
	return func(ac AuthContext, trc *Trace, left string, right policy.Value) bool {
		return !Cond_MatchAny(f)(ac, trc, left, right)
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

// Mod_Not defines a Condition modifier which flips the result of the underlying func
func Mod_Not(f Compare) Compare {
	return func(trc *Trace, left, right string) bool {
		return !f(trc, left, right)
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

// Mod_Date converts the string inputs to dates, allowing datewise comparisons
func Mod_Date(f func(*Trace, int, int) bool) Compare {
	return func(trc *Trace, left, right string) bool {
		tLeft, err := time.Parse(TIME_FORMAT, left)
		if err != nil {
			// TODO(nsiow) find a good place to log errors
			return false
		}
		nLeft := int(tLeft.Unix())

		rRight, err := time.Parse(TIME_FORMAT, left)
		if err != nil {
			// TODO(nsiow) find a good place to log errors
			return false
		}
		nRight := int(rRight.Unix())

		return f(trc, nLeft, nRight)
	}
}

// --------------------------------------------------------------------------------
// Externally facing functions
// --------------------------------------------------------------------------------

// ConditionResolveOperator takes in an operator name and resolves it to a function
//
// If the function could be resolved, the second return value is `true`. Otherwise, the second
// return value is `false`
func ConditionResolveOperator(op string) (ConditionOperator, bool) {
	// Handle function modifiers
	mods := []ConditionMod{}
	if strings.HasSuffix(op, "IfExists") {
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
	return f, true
}

// evalCondition is an evaluation helper function which performs a condition check over a single
// operation / key / value 3-tuple
func evalCondition(ac AuthContext, trc *Trace, f ConditionOperator, key string,
	right policy.Value) bool {
	left := ac.Key(key)
	// TODO(nsiow) `right` should get its policy variables expanded where relevant
	// except not here because it depends on the operator!
	return f(ac, trc, left, right)
}
