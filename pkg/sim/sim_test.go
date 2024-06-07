package sim

import (
	"fmt"
	"reflect"
	"testing"

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
	sim, err = NewSimulator(errorOpt)
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

// TestSimulateEvent validates the simulator's ability to correctly simulate a single event
//
// We are keeping these tests simple in terms of evaluation logic, as we really just want to test
// the simulator interface vs the logic which is tested deeply elsewhere
func TestSimulateEvent(t *testing.T) {
	type test struct {
		name  string
		event Event
		want  bool
		err   bool
	}

	tests := []test{
		{
			name: "same_account_implicit_deny",
			event: Event{
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
			want: false,
		},
		{
			name: "same_account_simple_allow",
			event: Event{
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
			want: true,
		},
	}

	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		sim, _ := NewSimulator()
		res, err := sim.SimulateEvent(&tc.event)

		check := true
		switch {
		case err == nil && tc.err:
			t.Fatalf("expected error, got success for test case '%s': %v", tc.name, err)
		case err != nil && tc.err:
			// expected error; got error
			t.Logf("test saw expected error: %v", err)
			check = false
		case err == nil && !tc.err:
			// no error and not expecting one, continue
		case err != nil && !tc.err:
			t.Fatalf("unable to create policy from test case '%s': %v", tc.name, err)
		}

		if check && !reflect.DeepEqual(res.IsAllowed, tc.want) {
			t.Fatalf("failed test case: '%s', wanted %v got %v", tc.name, tc.want, res.IsAllowed)
		}
	}
}

// TestSimulateByArn validates the simulator's ability to correctly simulate access based on ARN
// lookups
func TestSimulateByArn(t *testing.T) {
	type test struct {
		name         string
		env          *entities.Environment
		action       string
		principalArn string
		resourceArn  string
		want         bool
		err          bool
	}

	tests := []test{
		{
			name:         "test_allow",
			env:          &SimpleTestEnvironment_1,
			action:       "s3:listbucket",
			principalArn: "arn:aws:iam::88888:role/role1",
			resourceArn:  "arn:aws:s3:::bucket1",
			want:         true,
		},
		{
			name:         "test_deny",
			env:          &SimpleTestEnvironment_1,
			action:       "s3:listbucket",
			principalArn: "arn:aws:iam::88888:role/role1",
			resourceArn:  "arn:aws:s3:::bucket3",
			want:         false,
		},
		{
			name:         "test_empty_environment",
			env:          nil,
			action:       "s3:listbucket",
			principalArn: "arn:aws:iam::88888:role/role1",
			resourceArn:  "arn:aws:s3:::bucket1",
			err:          true,
		},
		{
			name:         "both_missing",
			env:          &SimpleTestEnvironment_1,
			action:       "s3:listbucket",
			principalArn: "arn:aws:iam::88888:role/doesnotexist",
			resourceArn:  "arn:aws:s3:::doesnotexist",
			err:          true,
		},
		{
			name:         "principal_missing",
			env:          &SimpleTestEnvironment_1,
			action:       "s3:listbucket",
			principalArn: "arn:aws:iam::88888:role/doesnotexist",
			resourceArn:  "arn:aws:s3:::bucket1",
			err:          true,
		},
		{
			name:         "resource_missing",
			env:          &SimpleTestEnvironment_1,
			action:       "s3:listbucket",
			principalArn: "arn:aws:iam::88888:role/role1",
			resourceArn:  "arn:aws:s3:::doesnotexist",
			err:          true,
		},
	}

	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		sim, _ := NewSimulator()
		sim.SetEnvironment(tc.env)
		res, err := sim.SimulateByArn(tc.action, tc.principalArn, tc.resourceArn, nil)

		check := true
		switch {
		case err == nil && tc.err:
			t.Fatalf("expected error, got success for test case '%s': %v", tc.name, err)
		case err != nil && tc.err:
			// expected error; got error
			t.Logf("test saw expected error: %v", err)
			check = false
		case err == nil && !tc.err:
			// no error and not expecting one, continue
		case err != nil && !tc.err:
			t.Fatalf("unable to create policy from test case '%s': %v", tc.name, err)
		}

		if check && !reflect.DeepEqual(res.IsAllowed, tc.want) {
			t.Fatalf("failed test case: '%s', wanted %v got %v", tc.name, tc.want, res.IsAllowed)
		}
	}
}

// TestComputeAccessSummary validates the simulator's ability to construct a summary of access
// between Principals + Resources
func TestComputeAccessSummary(t *testing.T) {
	type test struct {
		name    string
		env     *entities.Environment
		actions []string
		want    map[string]int
		err     bool
	}

	tests := []test{
		{
			name:    "simple_environment_1",
			env:     &SimpleTestEnvironment_1,
			actions: []string{"s3:listbucket"},
			want: map[string]int{
				"arn:aws:s3:::bucket1": 1,
				"arn:aws:s3:::bucket2": 1,
				"arn:aws:s3:::bucket3": 0,
			},
		},
		{
			name:    "unrelated_actions",
			env:     &SimpleTestEnvironment_1,
			actions: []string{"this:doesnotexist"},
			want: map[string]int{
				"arn:aws:s3:::bucket1": 0,
				"arn:aws:s3:::bucket2": 0,
				"arn:aws:s3:::bucket3": 0,
			},
		},
		{
			name: "empty_environment",
			env:  &entities.Environment{},
			want: map[string]int{},
		},
		{
			name: "no_environment",
			env:  nil,
			err:  true,
		},
		{
			name: "error_nonexistent_condition",
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
										"StringEqualsThisDoesNotExist": nil,
									},
								},
							},
						},
					},
				},
			},
			actions: []string{"s3:listbucket"},
			err:     true,
		},
	}

	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		sim, _ := NewSimulator(WithFailOnUnknownCondition())
		sim.SetEnvironment(tc.env)
		summary, err := sim.ComputeAccessSummary(tc.actions)

		check := true
		switch {
		case err == nil && tc.err:
			t.Fatalf("expected error, got success for test case '%s': %v", tc.name, err)
		case err != nil && tc.err:
			// expected error; got error
			t.Logf("test saw expected error: %v", err)
			check = false
		case err == nil && !tc.err:
			// no error and not expecting one, continue
		case err != nil && !tc.err:
			t.Fatalf("unable to create policy from test case '%s': %v", tc.name, err)
		}

		if check && !reflect.DeepEqual(summary, tc.want) {
			t.Fatalf("failed test case: '%s', wanted %v got %v", tc.name, tc.want, summary)
		}
	}
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
