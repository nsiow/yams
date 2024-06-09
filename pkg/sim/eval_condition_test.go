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
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		return evalStatementMatchesCondition(&i.options, i.ac, &Trace{}, &i.stmt)
	})
}

// TestStringEquals validates StringEquals/StringNotEquals behavior
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

// TestStringLike validates StringLike/StringNotEquals behavior
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
			Name: "simple_inverted:nomatch",
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
