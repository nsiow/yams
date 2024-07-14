package sim

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/nsiow/yams/internal/testrunner"
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
	sim, err = NewSimulator(WithFailOnUnknownCondition())
	if err != nil {
		t.Fatalf("unexpected error creating a simulator with options: %v", err)
	}
	if sim == nil {
		t.Fatalf("unexpected nil simulator when creating with options")
	}
	if sim.options.FailOnUnknownCondition != true {
		t.Fatalf("expected option FailOnUnknownCondition to be applied, but saw 'false'")
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

// TestSimulatorEnvironment validates our ability to manipulate the Environment of the simulator
func TestSimulatorEnvironment(t *testing.T) {

	// Define our environment
	env := entities.Environment{
		Principals: []entities.Principal{
			{
				Arn: "arn:aws:iam::88888:role/exampleRole",
			},
		},
	}

	// Create a simulator and set Environment
	sim, _ := NewSimulator()
	sim.SetEnvironment(&env)

	// Compare retrieved environment to ours
	got := sim.Environment()
	if !reflect.DeepEqual(env, *got) {
		t.Fatalf("retrieved environment %+v does not match ours: %+v", got, env)
	}
}

// TestSimulate validates the simulator's ability to correctly simulate a single event
//
// We are keeping these tests simple in terms of evaluation logic, as we really just want to test
// the simulator interface vs the logic which is tested deeply elsewhere
func TestSimulate(t *testing.T) {
	tests := []testrunner.TestCase[AuthContext, bool]{
		{
			Name: "same_account_implicit_deny",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn:              "arn:aws:iam::88888:role/myrole",
					Account:          "88888",
					InlinePolicies:   nil,
					AttachedPolicies: nil,
				},
				Resource: &entities.Resource{
					Arn:     "arn:aws:s3:::mybucket",
					Account: "88888",
					Policy:  policy.Policy{},
				},
			},
			Want: false,
		},
		{
			Name: "same_account_simple_allow",
			Input: AuthContext{
				Action: "s3:listbucket",
				Principal: &entities.Principal{
					Arn:     "arn:aws:iam::88888:role/myrole",
					Account: "88888",
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
					Arn:     "arn:aws:s3:::mybucket",
					Account: "88888",
				},
			},
			Want: true,
		},
	}

	testrunner.RunTestSuite(t, tests, func(ac AuthContext) (bool, error) {
		sim, _ := NewSimulator()
		res, err := sim.Simulate(ac)
		if err != nil {
			return false, err
		}

		return res.IsAllowed, nil
	})
}

// TestSimulateByArn validates the simulator's ability to correctly simulate access based on ARN
// lookups
func TestSimulateByArn(t *testing.T) {
	type input struct {
		env          *entities.Environment
		action       string
		principalArn string
		resourceArn  string
	}

	tests := []testrunner.TestCase[input, bool]{
		{
			Name: "test_allow",
			Input: input{
				env:          &SimpleTestEnvironment_1,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/role1",
				resourceArn:  "arn:aws:s3:::bucket1",
			},
			Want: true,
		},
		{
			Name: "test_deny",
			Input: input{
				env:          &SimpleTestEnvironment_1,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/role1",
				resourceArn:  "arn:aws:s3:::bucket3",
			},
			Want: false,
		},
		{
			Name: "test_empty_environment",
			Input: input{
				env:          nil,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/role1",
				resourceArn:  "arn:aws:s3:::bucket1",
			},
			ShouldErr: true,
		},
		{
			Name: "both_missing",
			Input: input{
				env:          &SimpleTestEnvironment_1,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/doesnotexist",
				resourceArn:  "arn:aws:s3:::doesnotexist",
			},
			ShouldErr: true,
		},
		{
			Name: "principal_missing",
			Input: input{
				env:          &SimpleTestEnvironment_1,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/doesnotexist",
				resourceArn:  "arn:aws:s3:::bucket1",
			},
			ShouldErr: true,
		},
		{
			Name: "resource_missing",
			Input: input{
				env:          &SimpleTestEnvironment_1,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/role1",
				resourceArn:  "arn:aws:s3:::doesnotexist",
			},
			ShouldErr: true,
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (bool, error) {
		sim, _ := NewSimulator()
		sim.SetEnvironment(i.env)
		res, err := sim.SimulateByArn(i.action, i.principalArn, i.resourceArn, nil)
		if err != nil {
			return false, err
		}

		return res.IsAllowed, nil
	})
}

// TestComputeAccessSummary validates the simulator's ability to construct a summary of access
// between Principals + Resources
func TestComputeAccessSummary(t *testing.T) {
	type input struct {
		env     *entities.Environment
		actions []string
	}

	tests := []testrunner.TestCase[input, map[string]int]{
		{
			Name: "simple_environment_1",
			Input: input{
				env:     &SimpleTestEnvironment_1,
				actions: []string{"s3:listbucket"},
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
				env:     &SimpleTestEnvironment_1,
				actions: []string{"this:doesnotexist"},
			},
			Want: map[string]int{
				"arn:aws:s3:::bucket1": 0,
				"arn:aws:s3:::bucket2": 0,
				"arn:aws:s3:::bucket3": 0,
			},
		},
		{
			Name: "empty_environment",
			Input: input{
				env: &entities.Environment{},
			},
			Want: map[string]int{},
		},
		{
			Name: "no_environment",
			Input: input{
				env: nil,
			},
			ShouldErr: true,
		},
		{
			Name:      "error_nonexistent_condition",
			ShouldErr: true,
			Input: input{
				actions: []string{"s3:listbucket"},
				env: &entities.Environment{
					Principals: []entities.Principal{
						{
							Arn:     "arn:aws:iam::88888:role/role1",
							Account: "88888",
						},
					},
					Resources: []entities.Resource{
						{
							Arn:     "arn:aws:s3:::mybucket",
							Account: "11111",
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

	testrunner.RunTestSuite(t, tests, func(i input) (map[string]int, error) {
		sim, _ := NewSimulator(WithFailOnUnknownCondition())
		sim.SetEnvironment(i.env)
		summary, err := sim.ComputeAccessSummary(i.actions)
		if err != nil {
			return nil, err
		}

		return summary, nil
	})
}

// SimpleTestEnvironment_1 defines a very simple but reusable test environment for basic,
// non-exhaustive unit tests
var SimpleTestEnvironment_1 entities.Environment = entities.Environment{
	Principals: []entities.Principal{
		{
			Arn:     "arn:aws:iam::88888:role/role1",
			Account: "88888",
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
			Arn:     "arn:aws:iam::88888:role/role2",
			Account: "88888",
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
			Arn:     "arn:aws:iam::88888:role/role3",
			Account: "11111",
		},
	},
	Resources: []entities.Resource{
		{
			Arn:     "arn:aws:s3:::bucket1",
			Account: "88888",
		},
		{
			Arn:     "arn:aws:s3:::bucket2",
			Account: "11111",
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
			Arn:     "arn:aws:s3:::bucket3",
			Account: "11111",
		},
	},
}
