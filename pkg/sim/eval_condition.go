package sim

import (
	"errors"
	"strings"

	"github.com/nsiow/yams/pkg/policy"
	"github.com/nsiow/yams/pkg/policy/condition"
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
type ConditionOperator = func(trc *Trace, left string, right policy.Value) bool

// ConditionMod defines a function which wraps a ConditionOperator
type ConditionMod = func(ConditionOperator) ConditionOperator

// --------------------------------------------------------------------------------
// Mappings
// --------------------------------------------------------------------------------

// ConditionOperatorMap defines the mapping between operator names and functions
var ConditionOperatorMap = map[string]ConditionOperator{
	condition.Op_StringEquals:    Cond_MatchAny(Cond_StringEquals),
	condition.Op_StringNotEquals: Cond_MatchNone(Cond_StringEquals),
	condition.Op_StringEqualsIgnoreCase: Cond_MatchAny(
		Mod_CaseInsensitive(Cond_StringEquals),
	),
	condition.Op_StringNotEqualsIgnoreCase: Cond_MatchNone(
		Mod_CaseInsensitive(Cond_StringEquals),
	),
}

// --------------------------------------------------------------------------------
// Condition evaluation functions
// --------------------------------------------------------------------------------

func Cond_MatchAny(f Compare) ConditionOperator {
	return func(trc *Trace, left string, right policy.Value) bool {
		for _, value := range right {
			if f(trc, left, value) {
				return true
			}
		}

		return false
	}
}

func Cond_MatchNone(f Compare) ConditionOperator {
	return func(trc *Trace, left string, right policy.Value) bool {
		return !Cond_MatchAny(f)(trc, left, right)
	}
}

// --------------------------------------------------------------------------------
// Condition comparison functions
// --------------------------------------------------------------------------------

// Cond_StringEquals defines the `StringEquals` condition function
func Cond_StringEquals(trc *Trace, left, right string) bool {
	return left == right
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

// Mod_MustExist defines a Condition modifier which returns false if the key is not found
func Mod_MustExist(f ConditionOperator) ConditionOperator {
	return func(trc *Trace, left string, right policy.Value) bool {
		if left == "" {
			return false
		}

		return f(trc, left, right)
	}
}

// Mod_IfExists defines a Condition modifier which returns true if the key is not found
func Mod_IfExists(f ConditionOperator) ConditionOperator {
	return func(trc *Trace, left string, right policy.Value) bool {
		if left == "" {
			return true
		}

		return f(trc, left, right)
	}
}

// Mod_CaseInsensitive defines a Condition modifier which ignores character casing
func Mod_CaseInsensitive(f Compare) Compare {
	return func(trc *Trace, left, right string) bool {
		return f(trc, strings.ToLower(left), strings.ToLower(right))
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

	// FIXME(nsiow) you are debugging this w.r.t. StringNotEquals and its incorrect behavior
	left := ac.Key(key)
	return f(trc, left, right)
}
