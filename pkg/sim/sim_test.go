package sim

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/aws/types"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// TestNewSimulator validates our ability to create new simulator with and without options
func TestNewSimulator(t *testing.T) {
	// Try with no options
	sim, err := NewSimulator()
	if err != nil {
		t.Fatalf("unexpected error creating a simulator with no options: %v", err)
	}
	if sim == nil {
		t.Fatalf("unexpected nil simulator when creating with no options")
	}

	// Try with a simple option; validate that it got applied
	sim, err = NewSimulator(WithSkipUnknownConditionOperators())
	if err != nil {
		t.Fatalf("unexpected error creating a simulator with options: %v", err)
	}
	if sim == nil {
		t.Fatalf("unexpected nil simulator when creating with options")
	}
	if sim.options.SkipUnknownConditionOperators != true {
		t.Fatalf("expected option SkipUnknownCondition to be applied, but saw 'false'")
	}

	// Try with an option that always fails
	errorOpt := func(opt *Options) error {
		return fmt.Errorf("expected error for testing")
	}
	_, err = NewSimulator(errorOpt)
	if err == nil {
		t.Fatalf("expected error with a custom option, but saw success")
	}
}

// TestSimulatorUniverse validates our ability to manipulate the Universe of the simulator
func TestSimulatorUniverse(t *testing.T) {

	// Define our universe
	universe := entities.Universe{
		Principals: []entities.Principal{
			{
				Arn: "arn:aws:iam::88888:role/exampleRole",
			},
		},
	}

	// Create a simulator and set Universe
	sim, _ := NewSimulator()
	sim.SetUniverse(universe)

	// Compare retrieved universe to ours
	got := sim.Universe()
	if !reflect.DeepEqual(universe, got) {
		t.Fatalf("retrieved universe %+v does not match ours: %+v", got, universe)
	}
}

// TestSimulate validates the simulator's ability to correctly simulate a single event
//
// We are keeping these tests simple in terms of evaluation logic, as we really just want to test
// the simulator interface vs the logic which is tested deeply elsewhere
func TestSimulate(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, bool]{
		{
			Name: "same_account_implicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.Principal{
					Arn:              "arn:aws:iam::88888:role/myrole",
					AccountId:        "88888",
					InlinePolicies:   nil,
					AttachedPolicies: nil,
				},
				Resource: &entities.Resource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
					Policy:    policy.Policy{},
				},
			},
			Want: false,
		},
		{
			Name: "same_account_simple_allow",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),
				Principal: &entities.Principal{
					Arn:       "arn:aws:iam::88888:role/myrole",
					AccountId: "88888",
					InlinePolicies: []policy.Policy{
						{
							Statement: []policy.Statement{
								{
									Effect:   policy.EFFECT_ALLOW,
									Action:   []string{"s3:listbucket"},
									Resource: []string{"arn:aws:s3:::mybucket"},
								},
							},
						},
					},
				},
				Resource: &entities.Resource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(ac AuthContext) (bool, error) {
		sim, _ := NewSimulator()
		res, err := sim.Simulate(ac)
		if err != nil {
			return false, err
		}

		t.Log(res.Trace.Log())
		return res.IsAllowed, nil
	})
}

// TestSimulateByArn validates the simulator's ability to correctly simulate access based on ARN
// lookups
func TestSimulateByArn(t *testing.T) {
	type input struct {
		universe     entities.Universe
		action       string
		principalArn string
		resourceArn  string
	}

	tests := []testlib.TestCase[input, bool]{
		{
			Name: "test_allow",
			Input: input{
				universe:     SimpleTestUniverse_1,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/role1",
				resourceArn:  "arn:aws:s3:::bucket1",
			},
			Want: true,
		},
		{
			Name: "test_deny",
			Input: input{
				universe:     SimpleTestUniverse_1,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/role1",
				resourceArn:  "arn:aws:s3:::bucket3",
			},
			Want: false,
		},
		{
			Name: "test_empty_universe",
			Input: input{
				universe:     entities.Universe{},
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/role1",
				resourceArn:  "arn:aws:s3:::bucket1",
			},
			ShouldErr: true,
		},
		{
			Name: "both_missing",
			Input: input{
				universe:     SimpleTestUniverse_1,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/doesnotexist",
				resourceArn:  "arn:aws:s3:::doesnotexist",
			},
			ShouldErr: true,
		},
		{
			Name: "principal_missing",
			Input: input{
				universe:     SimpleTestUniverse_1,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/doesnotexist",
				resourceArn:  "arn:aws:s3:::bucket1",
			},
			ShouldErr: true,
		},
		{
			Name: "resource_missing",
			Input: input{
				universe:     SimpleTestUniverse_1,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/role1",
				resourceArn:  "arn:aws:s3:::doesnotexist",
			},
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		sim, _ := NewSimulator()
		sim.SetUniverse(i.universe)
		res, err := sim.SimulateByArn(i.action, i.principalArn, i.resourceArn, nil)
		if err != nil {
			return false, err
		}

		t.Log(res.Trace.Log())
		return res.IsAllowed, nil
	})
}

// TestComputeAccessSummary validates the simulator's ability to construct a summary of access
// between Principals + Resources
func TestComputeAccessSummary(t *testing.T) {
	type input struct {
		universe entities.Universe
		actions  []types.Action
	}

	tests := []testlib.TestCase[input, map[string]int]{
		{
			Name: "simple_universe_1",
			Input: input{
				universe: SimpleTestUniverse_1,
				actions:  []types.Action{sar.MustLookupString("s3:listbucket")},
			},
			Want: map[string]int{
				"arn:aws:s3:::bucket1": 1,
				"arn:aws:s3:::bucket2": 1,
				"arn:aws:s3:::bucket3": 0,
			},
		},
		{
			Name: "unrelated_actions",
			Input: input{
				universe: SimpleTestUniverse_1,
				actions:  []types.Action{sar.MustLookupString("sqs:listqueues")},
			},
			Want: map[string]int{
				"arn:aws:s3:::bucket1": 0,
				"arn:aws:s3:::bucket2": 0,
				"arn:aws:s3:::bucket3": 0,
			},
		},
		{
			Name: "empty_universe",
			Input: input{
				universe: entities.Universe{},
			},
			Want: map[string]int{},
		},
		{
			Name:      "error_nonexistent_condition",
			ShouldErr: true,
			Input: input{
				actions: []types.Action{sar.MustLookupString("s3:listbucket")},
				universe: entities.Universe{
					Principals: []entities.Principal{
						{
							Arn:       "arn:aws:iam::88888:role/role1",
							AccountId: "88888",
						},
					},
					Resources: []entities.Resource{
						{
							Arn:       "arn:aws:s3:::mybucket",
							AccountId: "11111",
							Policy: policy.Policy{
								Statement: []policy.Statement{
									{
										Effect:   policy.EFFECT_ALLOW,
										Action:   []string{"s3:listbucket"},
										Resource: []string{"arn:aws:s3:::mybucket"},
										Principal: policy.Principal{
											AWS: []string{"arn:aws:iam::88888:role/role1"},
										},
										Condition: map[string]map[string]policy.Value{
											"StringEqualsThisDoesNotExist": {
												"foo": []string{"bar"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (map[string]int, error) {
		sim, _ := NewSimulator()
		sim.SetUniverse(i.universe)
		summary, err := sim.ComputeAccessSummary(i.actions)
		if err != nil {
			return nil, err
		}

		return summary, nil
	})
}

// SimpleTestUniverse_1 defines a very simple but reusable test universe for basic,
// non-exhaustive unit tests
var SimpleTestUniverse_1 = entities.Universe{
	Principals: []entities.Principal{
		{
			Arn:       "arn:aws:iam::88888:role/role1",
			AccountId: "88888",
			InlinePolicies: []policy.Policy{
				{
					Statement: []policy.Statement{
						{
							Effect:   policy.EFFECT_ALLOW,
							Action:   []string{"s3:listbucket"},
							Resource: []string{"*"},
						},
					},
				},
			},
		},
		{
			Arn:       "arn:aws:iam::88888:role/role2",
			AccountId: "88888",
			InlinePolicies: []policy.Policy{
				{
					Statement: []policy.Statement{
						{
							Effect:   policy.EFFECT_ALLOW,
							Action:   []string{"s3:listbucket"},
							Resource: []string{"arn:aws:s3:::bucket2"},
						},
					},
				},
			},
		},
		{
			Arn:       "arn:aws:iam::88888:role/role3",
			AccountId: "11111",
		},
	},
	Resources: []entities.Resource{
		{
			Arn:       "arn:aws:s3:::bucket1",
			AccountId: "88888",
		},
		{
			Arn:       "arn:aws:s3:::bucket2",
			AccountId: "11111",
			Policy: policy.Policy{
				Statement: []policy.Statement{
					{
						Effect:   policy.EFFECT_ALLOW,
						Action:   []string{"s3:listbucket"},
						Resource: []string{"arn:aws:s3:::mybucket"},
						Principal: policy.Principal{
							AWS: []string{"arn:aws:iam::88888:role/role2"},
						},
					},
				},
			},
		},
		{
			Arn:       "arn:aws:s3:::bucket3",
			AccountId: "11111",
		},
	},
}
