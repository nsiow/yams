package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/policy"
)

func TestEvalCheckCondition(t *testing.T) {
	type input struct {
		ac   *AuthContext
		opts *Options
		op   string
		cond policy.ConditionValues
	}

	tests := []testlib.TestCase[input, bool]{
		{
			Name: "simple_true_condition",
			Input: input{
				ac: &AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"color": "red",
					}),
				},
				op: "StringEquals",
				cond: policy.ConditionValues{
					"color": []string{"red"},
				},
			},
			Want: true,
		},
		{
			Name: "simple_false_condition",
			Input: input{
				ac: &AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"color": "blue",
					}),
				},
				op: "StringEquals",
				cond: policy.ConditionValues{
					"color": []string{"red"},
				},
			},
			Want: false,
		},
		{
			Name: "multi_true_condition",
			Input: input{
				ac: &AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"color": "red",
					}),
				},
				op: "StringEquals",
				cond: policy.ConditionValues{
					"color": []string{"red", "blue", "green"},
				},
			},
			Want: true,
		},
		{
			Name: "multi_false_condition",
			Input: input{
				ac: &AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"color": "red",
					}),
				},
				op: "StringEquals",
				cond: policy.ConditionValues{
					"color": []string{"yellow", "blue", "green"},
				},
			},
			Want: false,
		},
		{
			Name: "soft_fail_unknown_operator",
			Input: input{
				ac: &AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"color": "red",
					}),
				},
				op: "SomeOperator",
				cond: policy.ConditionValues{
					"color": []string{"red"},
				},
				opts: NewOptions(),
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		// Set default testing options if none provided
		if i.opts == nil {
			i.opts = TestingSimulationOptions
		}

		subj := newSubject(i.ac, i.opts)
		return evalCheckCondition(subj, i.op, i.cond), nil
	})
}
