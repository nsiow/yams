package sim

import (
	"testing"

	"github.com/nsiow/yams/pkg/policy"
)

func TestMod_MustExist(t *testing.T) {
	// Create a base function that always returns true
	baseFunc := func(s *subject, left string, right policy.Value) bool {
		return true
	}

	modFunc := Mod_MustExist(baseFunc)
	subj := &subject{}

	// Test with empty left - should return false
	result := modFunc(subj, "", policy.Value{"test"})
	if result {
		t.Fatal("Mod_MustExist should return false for empty left value")
	}

	// Test with non-empty left - should call base function
	result = modFunc(subj, "value", policy.Value{"test"})
	if !result {
		t.Fatal("Mod_MustExist should return true when left is non-empty and base returns true")
	}

	// Test with non-empty left and base that returns false
	falseFunc := func(s *subject, left string, right policy.Value) bool {
		return false
	}
	modFalseFunc := Mod_MustExist(falseFunc)
	result = modFalseFunc(subj, "value", policy.Value{"test"})
	if result {
		t.Fatal("Mod_MustExist should return false when base returns false")
	}
}

func TestMod_ForAllValues_EmptyMultiKey(t *testing.T) {
	// Test ForAllValues when MultiKey returns empty but ConditionKey has value
	baseFunc := func(s *subject, left string, right policy.Value) bool {
		return left == "expected"
	}

	modFunc := Mod_ForAllValues(baseFunc)

	// Create an auth context with single value but no multi-value
	ac := AuthContext{
		Properties: NewBagFromMap(map[string]string{"testkey": "expected"}),
	}
	opts := Options{SkipServiceAuthorizationValidation: true}
	subj := &subject{auth: ac, opts: opts}

	result := modFunc(subj, "testkey", policy.Value{"expected"})
	if !result {
		t.Fatal("Mod_ForAllValues should use single value when multi-key is empty")
	}
}

func TestMod_ForAnyValues_EmptyMultiKey(t *testing.T) {
	// Test ForAnyValues when MultiKey returns empty but ConditionKey has value
	baseFunc := func(s *subject, left string, right policy.Value) bool {
		return left == "expected"
	}

	modFunc := Mod_ForAnyValues(baseFunc)

	// Create an auth context with single value but no multi-value
	ac := AuthContext{
		Properties: NewBagFromMap(map[string]string{"testkey": "expected"}),
	}
	opts := Options{SkipServiceAuthorizationValidation: true}
	subj := &subject{auth: ac, opts: opts}

	result := modFunc(subj, "testkey", policy.Value{"expected"})
	if !result {
		t.Fatal("Mod_ForAnyValues should use single value when multi-key is empty")
	}
}

func TestMod_ForAllValues_AllMatch(t *testing.T) {
	baseFunc := func(s *subject, left string, right policy.Value) bool {
		for _, v := range right {
			if left == v {
				return true
			}
		}
		return false
	}

	modFunc := Mod_ForAllValues(baseFunc)

	ac := AuthContext{
		MultiValueProperties: NewBagFromMap(map[string][]string{"testkey": {"a", "b"}}),
	}
	opts := Options{SkipServiceAuthorizationValidation: true}
	subj := &subject{auth: ac, opts: opts}

	// All values match
	result := modFunc(subj, "testkey", policy.Value{"a", "b", "c"})
	if !result {
		t.Fatal("Mod_ForAllValues should return true when all values match")
	}

	// Not all values match
	result = modFunc(subj, "testkey", policy.Value{"a"}) // 'b' won't match
	if result {
		t.Fatal("Mod_ForAllValues should return false when not all values match")
	}
}

func TestMod_ForAnyValues_AnyMatch(t *testing.T) {
	baseFunc := func(s *subject, left string, right policy.Value) bool {
		for _, v := range right {
			if left == v {
				return true
			}
		}
		return false
	}

	modFunc := Mod_ForAnyValues(baseFunc)

	ac := AuthContext{
		MultiValueProperties: NewBagFromMap(map[string][]string{"testkey": {"a", "b"}}),
	}
	opts := Options{SkipServiceAuthorizationValidation: true}
	subj := &subject{auth: ac, opts: opts}

	// Any value matches
	result := modFunc(subj, "testkey", policy.Value{"a", "z"})
	if !result {
		t.Fatal("Mod_ForAnyValues should return true when any value matches")
	}

	// No values match
	result = modFunc(subj, "testkey", policy.Value{"x", "y", "z"})
	if result {
		t.Fatal("Mod_ForAnyValues should return false when no values match")
	}
}
