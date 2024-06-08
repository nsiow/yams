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

// ConditionOperator defines the shape of a Condition operator function
//
// The function should take in two strings where `left` is the observed value and `right` is what
// we are trying to match against
type ConditionOperator = func(trc *Trace, left string, right string) bool

// ConditionMod defines a function which wraps a ConditionOperator
type ConditionMod = func(ConditionOperator) ConditionOperator

// --------------------------------------------------------------------------------
// Mappings
// --------------------------------------------------------------------------------

// ConditionOperatorMap defines the mapping between operator names and functions
var ConditionOperatorMap = map[string]ConditionOperator{
	condition.Op_StringEquals:              Cond_StringEquals,
	condition.Op_StringNotEquals:           Mod_Not(Cond_StringEquals),
	condition.Op_StringEqualsIgnoreCase:    Mod_CaseInsensitive(Cond_StringEquals),
	condition.Op_StringNotEqualsIgnoreCase: Mod_CaseInsensitive(Mod_Not(Cond_StringEquals)),
}

// --------------------------------------------------------------------------------
// Condition operators
// --------------------------------------------------------------------------------

// Cond_StringEquals defines the `StringEquals` condition function
// TODO(nsiow) determine if trace should get passed all the way down here
func Cond_StringEquals(trc *Trace, left, right string) bool {
	return left == right
}

// --------------------------------------------------------------------------------
// Condition modifiers
// --------------------------------------------------------------------------------

// Mod_Not defines a Condition modifier which flips the result of the underlying func
func Mod_Not(f ConditionOperator) ConditionOperator {
	return func(trc *Trace, left, right string) bool {
		return !f(trc, left, right)
	}
}

// Mod_MustExist defines a Condition modifier which returns false if the key is not found
func Mod_MustExist(f ConditionOperator) ConditionOperator {
	return func(trc *Trace, left, right string) bool {
		if left == "" {
			return false
		}

		return f(trc, left, right)
	}
}

// Mod_IfExists defines a Condition modifier which returns true if the key is not found
func Mod_IfExists(f ConditionOperator) ConditionOperator {
	return func(trc *Trace, left, right string) bool {
		if left == "" {
			return true
		}

		return f(trc, left, right)
	}
}

// Mod_CaseInsensitive defines a Condition modifier which ignores character casing
func Mod_CaseInsensitive(f ConditionOperator) ConditionOperator {
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
	values policy.Value) bool {

	// FIXME(nsiow) you are debugging this w.r.t. StringNotEquals and its incorrect behavior
	left := ac.Key(key)
	for _, right := range values {
		isTrue := f(trc, left, right)
		if isTrue {
			return true
		}
	}

	return false
}
