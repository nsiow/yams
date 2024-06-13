package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testrunner"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

type input struct {
	ac      AuthContext
	stmt    policy.Statement
	options Options
}

// TestStatementBase checks some basic condition shape/matching logic
func TestStatementBase(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "empty_condition",
			Input: input{
				ac: AuthContext{},
				stmt: policy.Statement{
					Condition: nil,
				},
			},
			Want: true,
		},
		{
			Name: "half_empty_condition",
			Input: input{
				ac: AuthContext{},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringEquals": nil,
					},
				},
			},
			Want: false,
		},
		{
			Name: "nonexistent_operator_fail",
			Input: input{
				ac:      AuthContext{},
				options: Options{FailOnUnknownCondition: true},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringEqualsThisDoesNotExist": {
							"foo": []string{"bar"},
						},
					},
				},
			},
			ShouldErr: true,
		},
		{
			Name: "nonexistent_operator_no_fail",
			Input: input{
				ac:      AuthContext{},
				options: Options{FailOnUnknownCondition: false},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringEqualsThisDoesNotExist": {
							"foo": []string{"bar"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "nonexistent_lhs",
			Input: input{
				ac:      AuthContext{},
				options: Options{FailOnUnknownCondition: false},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringLike": {
							"": []string{"bar"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestStringEquals validates StringEquals behavior
func TestStringEquals(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Resource: &entities.Resource{Account: "55555"},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringEquals": {
							"aws:ResourceAccount": []string{"55555", "12345"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_nomatch",
			Input: input{
				ac: AuthContext{
					Resource: &entities.Resource{Account: "55555"},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringEquals": {
							"aws:ResourceAccount": []string{"77777"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "partial_match",
			Input: input{
				ac: AuthContext{
					Principal: &entities.Principal{Account: "12345"},
					Resource:  &entities.Resource{Account: "55555"},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringEquals": {
							"aws:PrincipalAccount": []string{"55555"},
							"aws:ResourceAccount":  []string{"55555", "12345"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestStringEqualsIgnoreCase validates StringEqualsIgnoreCase behavior
func TestStringEqualsIgnoreCase(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "ignorecase_match",
			Input: input{
				ac: AuthContext{
					Principal: &entities.Principal{Arn: "arn:aws:iam::55555:role/myrole"},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringEqualsIgnoreCase": {
							"aws:PrincipalArn": []string{"foo", "arn:aws:iam::55555:role/mYrOlE"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "ignorecase_no_match",
			Input: input{
				ac: AuthContext{
					Principal: &entities.Principal{Arn: "arn:aws:iam::55555:role/myrole"},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringEqualsIgnoreCase": {
							"aws:PrincipalArn": []string{"arn:aws:iam::55555:role/myrolee"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestStringNotEquals validates StringNotEquals behavior
func TestStringNotEquals(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "simple_inverted_match",
			Input: input{
				ac: AuthContext{
					Resource: &entities.Resource{Account: "55555"},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringNotEquals": {
							"aws:ResourceAccount": []string{"88888", "12345"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_inverted_nomatch",
			Input: input{
				ac: AuthContext{
					Resource: &entities.Resource{Account: "55555"},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringNotEquals": {
							"aws:ResourceAccount": []string{"55555", "12345"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestStringNotEqualsIgnoreCase validates StringNotEqualsIgnoreCase behavior
func TestStringNotEqualsIgnoreCase(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "ignorecase_match",
			Input: input{
				ac: AuthContext{
					Principal: &entities.Principal{Arn: "arn:aws:iam::55555:role/myrole"},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringNotEqualsIgnoreCase": {
							"aws:PrincipalArn": []string{"foo", "arn:aws:iam::55555:role/mYrOlE"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "ignorecase_no_match",
			Input: input{
				ac: AuthContext{
					Principal: &entities.Principal{Arn: "arn:aws:iam::55555:role/myrole"},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringNotEqualsIgnoreCase": {
							"aws:PrincipalArn": []string{"arn:aws:iam::55555:role/myrolee"},
						},
					},
				},
			},
			Want: true,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestStringLike validates StringLike behavior
func TestStringLike(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Resource: &entities.Resource{Account: "55555"},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringLike": {
							"aws:ResourceAccount": []string{"555*", "12*45"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "partial_match",
			Input: input{
				ac: AuthContext{
					Principal: &entities.Principal{Account: "12345"},
					Resource:  &entities.Resource{Account: "55555"},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringLike": {
							"aws:PrincipalAccount": []string{"555*"},
							"aws:ResourceAccount":  []string{"555*", "12*45"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestStringNotLike validates StringNotLike behavior
func TestStringNotLike(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "simple_inverted_match",
			Input: input{
				ac: AuthContext{
					Resource: &entities.Resource{Account: "55555"},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringNotLike": {
							"aws:ResourceAccount": []string{"888*", "12*45"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_inverted_nomatch",
			Input: input{
				ac: AuthContext{
					Resource: &entities.Resource{Account: "55555"},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringNotLike": {
							"aws:ResourceAccount": []string{"555*", "12*45"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestNumericConversion validates correct behavior of casting condition values to numbers
func TestNumericConversion(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "non_numeric_lhs",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeNumericKey": "foo",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericEquals": {
							"aws:SomeNumericKey": []string{"1234"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "non_numeric_rhs",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeNumericKey": "123",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericEquals": {
							"aws:SomeNumericKey": []string{"foo"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})

}

// TestNumericEquals validates NumericEquals behavior
func TestNumericEquals(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeNumericKey": "100",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericEquals": {
							// TODO(nsiow) validate that this is correct behavior for multivalue keys
							"aws:SomeNumericKey": []string{"123", "100"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_nomatch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeNumericKey": "100",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericEquals": {
							"aws:SomeNumericKey": []string{"500"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestNumericNotEquals validates NumericEquals behavior
func TestNumericNotEquals(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeNumericKey": "100",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericNotEquals": {
							"aws:SomeNumericKey": []string{"500"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_nomatch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeNumericKey": "100",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericNotEquals": {
							// TODO(nsiow) validate that this is correct behavior for multivalue keys
							"aws:SomeNumericKey": []string{"123", "100"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestNumericLessThan validates NumericLessThan behavior
func TestNumericLessThan(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeNumericKey": "100",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericLessThan": {
							// TODO(nsiow) validate that this is correct behavior for multivalue keys
							"aws:SomeNumericKey": []string{"150", "50"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_nomatch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeNumericKey": "100",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericLessThan": {
							"aws:SomeNumericKey": []string{"50"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestNumericLessThanEquals validates NumericLessThanEquals behavior
func TestNumericLessThanEquals(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeNumericKey": "100",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericLessThanEquals": {
							// TODO(nsiow) validate that this is correct behavior for multivalue keys
							"aws:SomeNumericKey": []string{"150", "50"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_equals_match",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeNumericKey": "100",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericLessThanEquals": {
							"aws:SomeNumericKey": []string{"100", "50"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_nomatch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeNumericKey": "100",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericLessThanEquals": {
							"aws:SomeNumericKey": []string{"50"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestNumericGreaterThan validates NumericGreaterThan behavior
func TestNumericGreaterThan(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "simple_nomatch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeNumericKey": "100",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericGreaterThan": {
							// TODO(nsiow) validate that this is correct behavior for multivalue keys
							"aws:SomeNumericKey": []string{"150", "50"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeNumericKey": "100",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericGreaterThan": {
							"aws:SomeNumericKey": []string{"50"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_nomatch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeNumericKey": "100",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericGreaterThan": {
							"aws:SomeNumericKey": []string{"150"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestNumericGreaterThanEquals validates NumericGreaterThanEquals behavior
func TestNumericGreaterThanEquals(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeNumericKey": "100",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericGreaterThanEquals": {
							// TODO(nsiow) validate that this is correct behavior for multivalue keys
							"aws:SomeNumericKey": []string{"150", "50"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_match_equals",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeNumericKey": "100",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericGreaterThanEquals": {
							"aws:SomeNumericKey": []string{"100", "50"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_nomatch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeNumericKey": "100",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericGreaterThanEquals": {
							"aws:SomeNumericKey": []string{"150"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestDateEquals validates DateEquals behavior
func TestDateEquals(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:11:12",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateEquals": {
							"aws:SomeDateKey": []string{"2023-01-01T03:04:05", "2024-01-01T10:11:12"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_match_epoch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "1704103872",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateEquals": {
							"aws:SomeDateKey": []string{"2023-01-01T03:04:05", "2024-01-01T10:11:12"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_nomatch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:11:12",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateEquals": {
							"aws:SomeDateKey": []string{"2024-01-01T10:12:11"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "invalid_lhs",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "foo",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateEquals": {
							"aws:SomeDateKey": []string{"2024-01-01T10:12:11"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "invalid_rhs",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:12:11",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateEquals": {
							"aws:SomeDateKey": []string{"foo"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestDateNotEquals validates DateNotEquals behavior
func TestDateNotEquals(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:11:12",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateNotEquals": {
							"aws:SomeDateKey": []string{"1212-12-12T12:12:12"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_nomatch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:11:12",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateNotEquals": {
							"aws:SomeDateKey": []string{"2023-01-01T03:04:05", "2024-01-01T10:11:12"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "simple_nomatch_epoch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "1704103872",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateNotEquals": {
							"aws:SomeDateKey": []string{"2023-01-01T03:04:05", "2024-01-01T10:11:12"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestDateLessThan validates DateLessThan behavior
func TestDateLessThan(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "1212-12-12T12:12:12",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateLessThan": {
							"aws:SomeDateKey": []string{"2024-01-01T10:11:12"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_nomatch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:11:12",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateLessThan": {
							"aws:SomeDateKey": []string{"2023-01-01T03:04:05", "1212-12-12T12:12:12"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "simple_match_epoch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "1704103872",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateLessThan": {
							"aws:SomeDateKey": []string{"2025-01-01T03:04:05"},
						},
					},
				},
			},
			Want: true,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestDateLessThanEquals validates DateLessThanEquals behavior
func TestDateLessThanEquals(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "1212-12-12T12:12:12",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateLessThanEquals": {
							"aws:SomeDateKey": []string{"2024-01-01T10:11:12"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_nomatch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:11:12",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateLessThanEquals": {
							"aws:SomeDateKey": []string{"2023-01-01T03:04:05", "1212-12-12T12:12:12"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "simple_match_epoch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "1704103872",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateLessThanEquals": {
							"aws:SomeDateKey": []string{"2025-01-01T03:04:05"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_equals_epoch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "1704103872",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateLessThanEquals": {
							"aws:SomeDateKey": []string{"1704103872"},
						},
					},
				},
			},
			Want: true,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestDateGreaterThan validates DateGreaterThan behavior
func TestDateGreaterThan(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:11:12",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateGreaterThan": {
							"aws:SomeDateKey": []string{"2023-01-01T03:04:05", "1212-12-12T12:12:12"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_nomatch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "1212-12-12T12:12:12",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateGreaterThan": {
							"aws:SomeDateKey": []string{"2024-01-01T10:11:12"},
						},
					},
				},
			},
			Want: false,
		},

		{
			Name: "simple_match_epoch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "1804103872",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateGreaterThan": {
							"aws:SomeDateKey": []string{"2025-01-01T03:04:05"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_equals_epoch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "1704103872",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateGreaterThan": {
							"aws:SomeDateKey": []string{"1704103872"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestDateGreaterThanEquals validates DateGreaterThanEquals behavior
func TestDateGreaterThanEquals(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:11:12",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateGreaterThanEquals": {
							"aws:SomeDateKey": []string{"2023-01-01T03:04:05", "1212-12-12T12:12:12"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_nomatch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "1212-12-12T12:12:12",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateGreaterThanEquals": {
							"aws:SomeDateKey": []string{"2024-01-01T10:11:12"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "simple_match_epoch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "1804103872",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateGreaterThanEquals": {
							"aws:SomeDateKey": []string{"2025-01-01T03:04:05"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_equals_epoch",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeDateKey": "1704103872",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateGreaterThanEquals": {
							"aws:SomeDateKey": []string{"1704103872"},
						},
					},
				},
			},
			Want: true,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestBool validates Bool behavior
func TestBool(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "simple_true",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SecureTransport": "true",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"Bool": {
							"aws:SecureTransport": []string{"true"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_false",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SecureTransport": "false",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"Bool": {
							"aws:SecureTransport": []string{"false"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_true_false",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SecureTransport": "true",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"Bool": {
							"aws:SecureTransport": []string{"false"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "simple_false_true",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SecureTransport": "false",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"Bool": {
							"aws:SecureTransport": []string{"true"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "ignore_case_true",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SecureTransport": "true",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"Bool": {
							"aws:SecureTransport": []string{"tRuE"},
						},
					},
				},
			},
			Want: true, // TODO(nsiow) validate that this is actually how Bool handles casing
		},
		{
			Name: "invalid_lhs",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SecureTransport": "foo",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"Bool": {
							"aws:SecureTransport": []string{"true"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "invalid_rhs",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SecureTransport": "true",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"Bool": {
							"aws:SecureTransport": []string{"foo"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestIfExists validates the ...IfExists behavior of condition operators
func TestIfExists(t *testing.T) {
	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "string_equals_if_exists",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeContextKey": "foo",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringEqualsIfExists": {
							"aws:SomeContextKey": []string{"foo"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "string_equals_if_exists_missing",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeContextKey": "foo",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringEqualsIfExists": {
							"aws:SomeOtherRandomDifferentContextKey": []string{"bar"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "numeric_equals_if_exists_missing",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeContextKey": "8888",
					},
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericEqualsIfExists": {
							"aws:SomeOtherRandomDifferentContextKey": []string{"1234"},
						},
					},
				},
			},
			Want: true,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}
