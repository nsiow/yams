package sim

import (
	"errors"
	"fmt"
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
type ConditionOperator = func(left string, right string) (bool, error)

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
func Cond_StringEquals(left, right string) (bool, error) {
	return left == right, nil
}

// --------------------------------------------------------------------------------
// Condition modifiers
// --------------------------------------------------------------------------------

// Mod_Not defines a Condition modifier which flips the result of the underlying func
func Mod_Not(f ConditionOperator) ConditionOperator {
	return func(left, right string) (bool, error) {
		x, err := f(left, right)
		return !x, err
	}
}

// Mod_MustExist defines a Condition modifier which returns false if the key is not found
func Mod_MustExist(f ConditionOperator) ConditionOperator {
	return func(left, right string) (bool, error) {
		if left == "" {
			return false, nil
		}

		return f(left, right)
	}
}

// Mod_IfExists defines a Condition modifier which returns true if the key is not found
func Mod_IfExists(f ConditionOperator) ConditionOperator {
	return func(left, right string) (bool, error) {
		if left == "" {
			return true, nil
		}

		return f(left, right)
	}
}

// Mod_CaseInsensitive defines a Condition modifier which ignores character casing
func Mod_CaseInsensitive(f ConditionOperator) ConditionOperator {
	return func(left, right string) (bool, error) {
		return f(strings.ToLower(left), strings.ToLower(right))
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
func evalCondition(ac AuthContext, op string, key string, values policy.Value) (bool, error) {
	f, exists := ConditionResolveOperator(op)
	if !exists {
		return false, fmt.Errorf("unknown operator '%s': %w", op, ErrUnknownOperator)
	}

	left := ac.Key(key)
	for _, right := range values {
		isTrue, err := f(left, right)
		if err != nil {
			// TODO(nsiow) add this to trace
			continue
		}

		if isTrue {
			return true, nil
		}
	}

	return false, nil
}
