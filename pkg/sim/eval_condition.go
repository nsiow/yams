package sim

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nsiow/yams/pkg/policy"
)

var ErrorUnknownOperator = errors.New("unknown operation")

type ConditionOperator = func(left string, right string) (bool, error)
type ConditionMod = func(ConditionOperator) ConditionOperator

var ConditionOperatorMap = map[string]ConditionOperator{
	"StringEquals":              Cond_StringEquals,
	"StringNotEquals":           Mod_Not(Cond_StringEquals),
	"StringEqualsIgnoreCase":    Mod_CaseInsensitive(Cond_StringEquals),
	"StringNotEqualsIgnoreCase": Mod_CaseInsensitive(Mod_Not(Cond_StringEquals)),
}

func Cond_StringEquals(left, right string) (bool, error) {
	return left == right, nil
}

func Mod_Not(f ConditionOperator) ConditionOperator {
	return func(left, right string) (bool, error) {
		x, err := f(left, right)
		return !x, err
	}
}

func Mod_MustExist(f ConditionOperator) ConditionOperator {
	return func(left, right string) (bool, error) {
		if left == "" {
			return false, nil
		}

		return f(left, right)
	}
}

func Mod_IfExists(f ConditionOperator) ConditionOperator {
	return func(left, right string) (bool, error) {
		if left == "" {
			return true, nil
		}

		return f(left, right)
	}
}

func Mod_CaseInsensitive(f ConditionOperator) ConditionOperator {
	return func(left, right string) (bool, error) {
		return f(strings.ToLower(left), strings.ToLower(right))
	}
}

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

func evalCondition(ac AuthContext, op string, key string, values policy.Value) (bool, error) {
	f, exists := ConditionResolveOperator(op)
	if !exists {
		return false, fmt.Errorf("unknown operator '%s': %w", op, ErrorUnknownOperator)
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
