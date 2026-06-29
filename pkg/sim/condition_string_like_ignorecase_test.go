package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/policy"
)

func TestStringLikeIgnoreCase(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "case_insensitive_wildcard_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:Foo": "ABCDEF",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringLikeIgnoreCase": {"aws:Foo": []string{"abc*"}},
					},
				},
			},
			Want: true,
		},
		{
			Name: "no_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:Foo": "ABCDEF",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringLikeIgnoreCase": {"aws:Foo": []string{"xyz*"}},
					},
				},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(&subj, &i.stmt), nil
	})
}

func TestStringNotLikeIgnoreCase(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "present_no_pattern_match_returns_true",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:Foo": "ABCDEF",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringNotLikeIgnoreCase": {"aws:Foo": []string{"xyz*"}},
					},
				},
			},
			Want: true,
		},
		{
			Name: "present_pattern_match_returns_false",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:Foo": "ABCDEF",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringNotLikeIgnoreCase": {"aws:Foo": []string{"abc*"}},
					},
				},
			},
			Want: false,
		},
		{
			Name: "missing_returns_true",
			Input: input{
				ac: AuthContext{},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"StringNotLikeIgnoreCase": {"aws:Missing": []string{"abc*"}},
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
