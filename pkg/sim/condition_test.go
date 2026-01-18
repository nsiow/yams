package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

type input struct {
	ac   AuthContext
	stmt policy.Statement
}

func TestStatementBase(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
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
			Name: "nonexistent_lhs",
			Input: input{
				ac: AuthContext{},
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
		{
			Name: "nonexistent_operator",
			Input: input{
				ac: AuthContext{},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringEqualsThisDoesNotExist": {
							"foo": []string{"bar"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestStringEquals(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Resource: &entities.FrozenResource{AccountId: "55555"},
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
					Resource: &entities.FrozenResource{AccountId: "55555"},
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
					Principal: &entities.FrozenPrincipal{AccountId: "12345"},
					Resource:  &entities.FrozenResource{AccountId: "55555"},
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

// --------------------------------------------------------------------------------
// String tests
// --------------------------------------------------------------------------------

func TestStringEqualsIgnoreCase(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "ignorecase_match",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						Arn: "arn:aws:iam::55555:role/myrole",
					},
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
					Principal: &entities.FrozenPrincipal{
						Arn: "arn:aws:iam::55555:role/myrole",
					},
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestStringNotEquals(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_inverted_match",
			Input: input{
				ac: AuthContext{
					Resource: &entities.FrozenResource{AccountId: "55555"},
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
					Resource: &entities.FrozenResource{AccountId: "55555"},
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestStringNotEqualsIgnoreCase(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "ignorecase_match",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						Arn: "arn:aws:iam::55555:role/myrole",
					},
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
					Principal: &entities.FrozenPrincipal{
						Arn: "arn:aws:iam::55555:role/myrole",
					},
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestStringLike(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Resource: &entities.FrozenResource{AccountId: "55555"},
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
					Principal: &entities.FrozenPrincipal{AccountId: "12345"},
					Resource:  &entities.FrozenResource{AccountId: "55555"},
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestStringNotLike(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_inverted_match",
			Input: input{
				ac: AuthContext{
					Resource: &entities.FrozenResource{AccountId: "55555"},
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
					Resource: &entities.FrozenResource{AccountId: "55555"},
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

// --------------------------------------------------------------------------------
// Numeric tests
// --------------------------------------------------------------------------------

func TestNumericConversion(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "non_numeric_lhs",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeNumericKey": "foo",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeNumericKey": "123",
					}),
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})

}

func TestNumericEquals(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeNumericKey": "100",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeNumericKey": "100",
					}),
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestNumericNotEquals(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeNumericKey": "100",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeNumericKey": "100",
					}),
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestNumericLessThan(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeNumericKey": "100",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeNumericKey": "100",
					}),
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestNumericLessThanEquals(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeNumericKey": "100",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeNumericKey": "100",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeNumericKey": "100",
					}),
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestNumericGreaterThan(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_nomatch",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeNumericKey": "100",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeNumericKey": "100",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeNumericKey": "100",
					}),
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

// --------------------------------------------------------------------------------
// Number tests
// --------------------------------------------------------------------------------

func TestNumericGreaterThanEquals(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeNumericKey": "100",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeNumericKey": "100",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeNumericKey": "100",
					}),
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

// --------------------------------------------------------------------------------
// Date tests
// --------------------------------------------------------------------------------

func TestDateEquals(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:11:12",
					}),
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
			Name: "simple_match_tz",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:11:12Z",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateEquals": {
							"aws:SomeDateKey": []string{"2023-01-01T03:04:05Z", "2024-01-01T10:11:12Z"},
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "1704103872",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:11:12",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "foo",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:12:11",
					}),
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestDateNotEquals(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:11:12",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:11:12",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "1704103872",
					}),
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestDateLessThan(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "1212-12-12T12:12:12",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:11:12",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "1704103872",
					}),
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestDateLessThanEquals(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "1212-12-12T12:12:12",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:11:12",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "1704103872",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "1704103872",
					}),
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestDateGreaterThan(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:11:12",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "1212-12-12T12:12:12",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "1804103872",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "1704103872",
					}),
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestDateGreaterThanEquals(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "2024-01-01T10:11:12",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "1212-12-12T12:12:12",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "1804103872",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeDateKey": "1704103872",
					}),
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

// --------------------------------------------------------------------------------
// Boolean tests
// --------------------------------------------------------------------------------

func TestBool(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_true",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SecureTransport": "true",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SecureTransport": "false",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SecureTransport": "true",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SecureTransport": "false",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SecureTransport": "true",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SecureTransport": "foo",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SecureTransport": "true",
					}),
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

// --------------------------------------------------------------------------------
// Binary tests
// --------------------------------------------------------------------------------

func TestBinary(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_true",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeBinaryKey": "Zm9vCg==",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"BinaryEquals": {
							"aws:SomeBinaryKey": []string{"Zm9vCg=="},
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeBinaryKey": "YmFyCg==",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"BinaryEquals": {
							"aws:SomeBinaryKey": []string{"Zm9vCg=="},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "equal_but_invalid",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeBinaryKey": "foo",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"BinaryEquals": {
							"aws:SomeBinaryKey": []string{"Zm9vCg=="},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "equal_but_invalid_reversed",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeBinaryKey": "Zm9vCg==",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"BinaryEquals": {
							"aws:SomeBinaryKey": []string{"foo"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

// --------------------------------------------------------------------------------
// IpAddress tests
// --------------------------------------------------------------------------------

func TestIpAddress(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIp": "10.0.0.1",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"IpAddress": {
							"aws:SourceIp": []string{"10.0.0.0/8"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_match_ipv6",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIp": "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"IpAddress": {
							"aws:SourceIp": []string{"2001:0db8:85a3::/64"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_match_multi",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIp": "10.0.0.1",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"IpAddress": {
							"aws:SourceIp": []string{"1.2.3.0/24", "10.0.0.0/8"},
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIp": "128.252.0.1",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"IpAddress": {
							"aws:SourceIp": []string{"10.0.0.0/8"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "simple_nomatch_ipv6",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIp": "2001:1db8:85a3:0000:0000:8a2e:0370:7334",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"IpAddress": {
							"aws:SourceIp": []string{"2001:0db8:85a3::/64"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "match_but_not_ips",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIp": "foo",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"IpAddress": {
							"aws:SourceIp": []string{"10.0.0.0/8"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "match_but_not_ips_reversed",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIp": "10.0.0.1",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"IpAddress": {
							"aws:SourceIp": []string{"foo"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "match_but_wrong_order",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIp": "10.0.0.0/8",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"IpAddress": {
							"aws:SourceIp": []string{"10.0.0.1"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestNotIpAddress(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_nomatch",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIp": "10.0.0.1",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NotIpAddress": {
							"aws:SourceIp": []string{"10.0.0.0/8"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "simple_nomatch_ipv6",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIp": "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NotIpAddress": {
							"aws:SourceIp": []string{"2001:0db8:85a3::/64"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "simple_nomatch_multi",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIp": "10.0.0.1",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NotIpAddress": {
							"aws:SourceIp": []string{"1.2.3.0/24", "10.0.0.0/8"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIp": "128.252.0.1",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NotIpAddress": {
							"aws:SourceIp": []string{"10.0.0.0/8"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_match_ipv6",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIp": "2001:1db8:85a3:0000:0000:8a2e:0370:7334",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NotIpAddress": {
							"aws:SourceIp": []string{"2001:0db8:85a3::/64"},
						},
					},
				},
			},
			Want: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

// --------------------------------------------------------------------------------
// Arn tests
// --------------------------------------------------------------------------------

func TestArnEquals(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceArn": "arn:aws:sns:us-east-1:88888:mytopic",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ArnEquals": {
							"aws:SourceArn": []string{"arn:aws:sns:us-east-1:88888:mytopic"},
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceArn": "arn:aws:sns:us-east-1:88888:mytopic",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ArnEquals": {
							"aws:SourceArn": []string{"arn:aws:sns:us-east-1:88888:othertopic"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "match_diff_region",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceArn": "arn:aws:sns:us-east-1:88888:mytopic",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ArnEquals": {
							"aws:SourceArn": []string{"arn:aws:sns:*:88888:mytopic"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "nomatch_diff_account",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceArn": "arn:aws:sns:us-east-1:88888:mytopic",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ArnEquals": {
							"aws:SourceArn": []string{"arn:aws:sns:us-east-1:*:othertopic"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestArnNotEquals(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_nomatch",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceArn": "arn:aws:sns:us-east-1:88888:mytopic",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ArnNotEquals": {
							"aws:SourceArn": []string{"arn:aws:sns:us-east-1:88888:mytopic"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceArn": "arn:aws:sns:us-east-1:88888:mytopic",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ArnNotEquals": {
							"aws:SourceArn": []string{"arn:aws:sns:us-east-1:88888:othertopic"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "nomatch_diff_region",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceArn": "arn:aws:sns:us-east-1:88888:mytopic",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ArnNotEquals": {
							"aws:SourceArn": []string{"arn:aws:sns:*:88888:mytopic"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "match_diff_account",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceArn": "arn:aws:sns:us-east-1:88888:mytopic",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ArnNotEquals": {
							"aws:SourceArn": []string{"arn:aws:sns:us-east-1:*:othertopic"},
						},
					},
				},
			},
			Want: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestArnLike(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceArn": "arn:aws:sns:us-east-1:88888:mytopic",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ArnLike": {
							"aws:SourceArn": []string{"arn:aws:sns:us-east-1:88888:mytopic"},
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceArn": "arn:aws:sns:us-east-1:88888:mytopic",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ArnLike": {
							"aws:SourceArn": []string{"arn:aws:sns:us-east-1:88888:othertopic"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "match_diff_region",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceArn": "arn:aws:sns:us-east-1:88888:mytopic",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ArnLike": {
							"aws:SourceArn": []string{"arn:aws:sns:*:88888:mytopic"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "nomatch_diff_account",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceArn": "arn:aws:sns:us-east-1:88888:mytopic",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ArnLike": {
							"aws:SourceArn": []string{"arn:aws:sns:us-east-1:*:othertopic"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestArnNotLike(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_nomatch",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceArn": "arn:aws:sns:us-east-1:88888:mytopic",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ArnNotLike": {
							"aws:SourceArn": []string{"arn:aws:sns:us-east-1:88888:mytopic"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "simple_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceArn": "arn:aws:sns:us-east-1:88888:mytopic",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ArnNotLike": {
							"aws:SourceArn": []string{"arn:aws:sns:us-east-1:88888:othertopic"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "nomatch_diff_region",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceArn": "arn:aws:sns:us-east-1:88888:mytopic",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ArnNotLike": {
							"aws:SourceArn": []string{"arn:aws:sns:*:88888:mytopic"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "match_diff_account",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceArn": "arn:aws:sns:us-east-1:88888:mytopic",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ArnNotLike": {
							"aws:SourceArn": []string{"arn:aws:sns:us-east-1:*:othertopic"},
						},
					},
				},
			},
			Want: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

// -----------------------------------------------------------------------------------------------
// Test weird stuff
// -----------------------------------------------------------------------------------------------

func TestIfExists(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "string_equals_if_exists",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeContextKey": "foo",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeContextKey": "foo",
					}),
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeContextKey": "8888",
					}),
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

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestForAllValues(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_equals",
			Input: input{
				ac: AuthContext{
					MultiValueProperties: NewBagFromMap(map[string][]string{
						"aws:TagKeys": {
							"foo", "bar", "baz",
						},
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ForAllValues:StringEquals": {
							"aws:TagKeys": []string{"foo", "bar", "baz"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_not_equals",
			Input: input{
				ac: AuthContext{
					MultiValueProperties: NewBagFromMap(map[string][]string{
						"aws:TagKeys": {
							"foo", "bar", "baz", "other",
						},
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ForAllValues:StringEquals": {
							"aws:TagKeys": []string{"foo", "bar", "baz"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "absent_key",
			Input: input{
				ac: AuthContext{},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ForAllValues:StringEquals": {
							"aws:SomeKey": []string{"foo"},
						},
					},
				},
			},
			Want: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

func TestForAnyValues(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_equals",
			Input: input{
				ac: AuthContext{
					MultiValueProperties: NewBagFromMap(map[string][]string{
						"aws:TagKeys": {
							"baz",
						},
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ForAnyValues:StringEquals": {
							"aws:TagKeys": []string{"foo", "bar", "baz"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "simple_not_equals",
			Input: input{
				ac: AuthContext{
					MultiValueProperties: NewBagFromMap(map[string][]string{
						"aws:TagKeys": {
							"lots", "of", "other", "strings",
						},
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ForAnyValues:StringEquals": {
							"aws:TagKeys": []string{"foo", "bar", "baz"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "absent_key",
			Input: input{
				ac: AuthContext{},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ForAnyValues:StringEquals": {
							"aws:SomeKey": []string{"foo"},
						},
					},
				},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}

// --------------------------------------------------------------------------------
// Null tests
// --------------------------------------------------------------------------------

func TestNull(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "key_absent_want_absent",
			Input: input{
				ac: AuthContext{},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"Null": {
							"aws:TokenIssueTime": []string{"true"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "key_present_want_absent",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:TokenIssueTime": "2024-01-01T00:00:00Z",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"Null": {
							"aws:TokenIssueTime": []string{"true"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "key_present_want_present",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:TokenIssueTime": "2024-01-01T00:00:00Z",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"Null": {
							"aws:TokenIssueTime": []string{"false"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "key_absent_want_present",
			Input: input{
				ac: AuthContext{},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"Null": {
							"aws:TokenIssueTime": []string{"false"},
						},
					},
				},
			},
			Want: false,
		},
		{
			Name: "case_insensitive_true",
			Input: input{
				ac: AuthContext{},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"Null": {
							"aws:TokenIssueTime": []string{"TRUE"},
						},
					},
				},
			},
			Want: true,
		},
		{
			Name: "case_insensitive_false",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:TokenIssueTime": "2024-01-01T00:00:00Z",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"Null": {
							"aws:TokenIssueTime": []string{"FALSE"},
						},
					},
				},
			},
			Want: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(subj, &i.stmt), nil
	})
}
