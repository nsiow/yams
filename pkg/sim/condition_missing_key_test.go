package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/policy"
)

// Negated operators must NOT match when the key is absent from the request context.
// Per AWS docs, callers must opt into "missing-as-true" behavior with the IfExists variant.
func TestNegatedOperators_MissingKey(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "string_not_equals",
			Input: input{
				ac: AuthContext{},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringNotEquals": {"aws:Missing": []string{"foo"}},
					},
				},
			},
			Want: false,
		},
		{
			Name: "string_not_equals_ignore_case",
			Input: input{
				ac: AuthContext{},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringNotEqualsIgnoreCase": {"aws:Missing": []string{"FOO"}},
					},
				},
			},
			Want: false,
		},
		{
			Name: "string_not_like",
			Input: input{
				ac: AuthContext{},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringNotLike": {"aws:Missing": []string{"foo*"}},
					},
				},
			},
			Want: false,
		},
		{
			Name: "numeric_not_equals",
			Input: input{
				ac: AuthContext{},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NumericNotEquals": {"aws:Missing": []string{"1"}},
					},
				},
			},
			Want: false,
		},
		{
			Name: "date_not_equals",
			Input: input{
				ac: AuthContext{},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"DateNotEquals": {"aws:Missing": []string{"2024-01-01"}},
					},
				},
			},
			Want: false,
		},
		{
			Name: "not_ip_address",
			Input: input{
				ac: AuthContext{},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"NotIpAddress": {"aws:Missing": []string{"10.0.0.0/8"}},
					},
				},
			},
			Want: false,
		},
		{
			Name: "arn_not_equals",
			Input: input{
				ac: AuthContext{},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ArnNotEquals": {"aws:Missing": []string{"arn:aws:iam::*:role/*"}},
					},
				},
			},
			Want: false,
		},
		{
			Name: "arn_not_like",
			Input: input{
				ac: AuthContext{},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"ArnNotLike": {"aws:Missing": []string{"arn:aws:iam::*:role/*"}},
					},
				},
			},
			Want: false,
		},
		// IfExists opts in to "missing-as-true" behavior.
		{
			Name: "string_not_equals_if_exists_missing",
			Input: input{
				ac: AuthContext{},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringNotEqualsIfExists": {"aws:Missing": []string{"foo"}},
					},
				},
			},
			Want: true,
		},
		// Negated operators with present, non-matching values still match.
		{
			Name: "string_not_equals_present_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:Foo": "actual",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringNotEquals": {"aws:Foo": []string{"different"}},
					},
				},
			},
			Want: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(&subj, &i.stmt), nil
	})
}
