package sim

import (
	"reflect"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/aws/sar/types"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

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
	sim, err = NewSimulator(WithSkipServiceAuthorizationValidation())
	if err != nil {
		t.Fatalf("unexpected error creating a simulator with options: %v", err)
	}
	if sim == nil {
		t.Fatalf("unexpected nil simulator when creating with options")
	}
	if sim.options.SkipServiceAuthorizationValidation != true {
		t.Fatalf("expected option SkipUnknownCondition to be applied, but saw 'false'")
	}
}

func TestSimulatorUniverse(t *testing.T) {

	// Define our uv
	uv := entities.NewUniverse()
	uv.PutAccount(entities.Account{Id: "55555"})

	// Create a simulator and set Universe
	sim, _ := NewSimulator()
	sim.SetUniverse(uv)

	// Compare retrieved uv to ours
	got := sim.Universe()
	if !reflect.DeepEqual(uv, got) {
		t.Fatalf("retrieved uv %+v does not match ours: %+v", got, uv)
	}
}

func TestSimulate(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, bool]{
		{
			Name: "same_account_implicit_deny",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:listbucket"),

				Principal: &entities.FrozenPrincipal{
					Arn:              "arn:aws:iam::88888:role/myrole",
					AccountId:        "88888",
					InlinePolicies:   nil,
					AttachedPolicies: nil,
				},
				Resource: &entities.FrozenResource{
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
				Principal: &entities.FrozenPrincipal{
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
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			Want: true,
		},
		{
			Name: "invalid_auth_context",
			Input: AuthContext{
				Action: sar.MustLookupString("sqs:getqueueurl"),
				Principal: &entities.FrozenPrincipal{
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
				Resource: &entities.FrozenResource{
					Arn:       "arn:aws:s3:::mybucket",
					AccountId: "88888",
				},
			},
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(ac AuthContext) (bool, error) {
		sim, _ := NewSimulator()
		res, err := sim.Simulate(ac)
		if err != nil {
			return false, err
		}

		return res.IsAllowed, nil
	})
}

func TestSimulateByArn(t *testing.T) {
	type input struct {
		uv           *entities.Universe
		action       string
		principalArn string
		resourceArn  string
	}

	tests := []testlib.TestCase[input, bool]{
		{
			Name: "test_allow",
			Input: input{
				uv:           SimpleTestUniverse_1,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/role1",
				resourceArn:  "arn:aws:s3:::bucket1",
			},
			Want: true,
		},
		{
			Name: "test_deny",
			Input: input{
				uv:           SimpleTestUniverse_1,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/role1",
				resourceArn:  "arn:aws:s3:::bucket3",
			},
			Want: false,
		},
		{
			Name: "test_empty_uv",
			Input: input{
				uv:           entities.NewUniverse(),
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/role1",
				resourceArn:  "arn:aws:s3:::bucket1",
			},
			ShouldErr: true,
		},
		{
			Name: "both_missing",
			Input: input{
				uv:           SimpleTestUniverse_1,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/doesnotexist",
				resourceArn:  "arn:aws:s3:::doesnotexist",
			},
			ShouldErr: true,
		},
		{
			Name: "principal_missing",
			Input: input{
				uv:           SimpleTestUniverse_1,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/doesnotexist",
				resourceArn:  "arn:aws:s3:::bucket1",
			},
			ShouldErr: true,
		},
		{
			Name: "resource_missing",
			Input: input{
				uv:           SimpleTestUniverse_1,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/role1",
				resourceArn:  "arn:aws:s3:::doesnotexist",
			},
			ShouldErr: true,
		},
		{
			Name: "invalid_action",
			Input: input{
				uv:           SimpleTestUniverse_1,
				action:       "s3:doesnotexist",
				principalArn: "arn:aws:iam::88888:role/role1",
				resourceArn:  "arn:aws:s3:::doesnotexist",
			},
			ShouldErr: true,
		},
		{
			Name: "cannot_freeze_principal",
			Input: input{
				uv:           InvalidTestUniverse_1,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/role1",
				resourceArn:  "arn:aws:s3:::bucket1",
			},
			ShouldErr: true,
		},
		{
			Name: "cannot_freeze_resources",
			Input: input{
				uv:           InvalidTestUniverse_2,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/role1",
				resourceArn:  "arn:aws:s3:::bucket1",
			},
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		sim, _ := NewSimulator()
		sim.SetUniverse(i.uv)
		res, err := sim.SimulateByArn(
			i.action,
			entities.Arn(i.principalArn),
			entities.Arn(i.resourceArn),
			nil,
		)
		if err != nil {
			return false, err
		}

		return res.IsAllowed, nil
	})

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		sim, _ := NewSimulator()
		sim.SetUniverse(i.uv)
		res, err := sim.SimulateByArnString(
			i.action,
			i.principalArn,
			i.resourceArn,
			nil,
		)
		if err != nil {
			return false, err
		}

		return res.IsAllowed, nil
	})
}

func TestComputeAccessSummary(t *testing.T) {
	type input struct {
		uv      *entities.Universe
		actions []*types.Action
	}

	tests := []testlib.TestCase[input, map[string]int]{
		{
			Name: "simple_uv_1",
			Input: input{
				uv:      SimpleTestUniverse_1,
				actions: []*types.Action{sar.MustLookupString("s3:listbucket")},
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
				uv:      SimpleTestUniverse_1,
				actions: []*types.Action{sar.MustLookupString("sns:publish")},
			},
			Want: map[string]int{
				"arn:aws:s3:::bucket1": 0,
				"arn:aws:s3:::bucket2": 0,
				"arn:aws:s3:::bucket3": 0,
			},
		},
		{
			Name: "empty_uv",
			Input: input{
				uv: entities.NewUniverse(),
			},
			Want: map[string]int{},
		},
		{
			Name: "cannot_freeze_principals",
			Input: input{
				uv: InvalidTestUniverse_1,
			},
			ShouldErr: true,
		},
		{
			Name: "cannot_freeze_resources",
			Input: input{
				uv: InvalidTestUniverse_2,
			},
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (map[string]int, error) {
		sim, _ := NewSimulator()
		sim.options = *TestingSimulationOptions
		sim.SetUniverse(i.uv)
		summary, err := sim.ComputeAccessSummary(i.actions)
		if err != nil {
			return nil, err
		}

		return summary, nil
	})
}

var SimpleTestUniverse_1 = entities.NewBuilder().
	WithPrincipals(
		entities.Principal{
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
		entities.Principal{
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
		entities.Principal{
			Arn:       "arn:aws:iam::88888:role/role3",
			AccountId: "11111",
		},
	).
	WithResources(
		entities.Resource{
			Arn:       "arn:aws:s3:::bucket1",
			AccountId: "88888",
		},
		entities.Resource{
			Arn:       "arn:aws:s3:::bucket2",
			AccountId: "11111",
			Policy: policy.Policy{
				Statement: []policy.Statement{
					{
						Effect:   policy.EFFECT_ALLOW,
						Action:   []string{"s3:listbucket"},
						Resource: []string{"arn:aws:s3:::bucket2"},
						Principal: policy.Principal{
							AWS: []string{"arn:aws:iam::88888:role/role2"},
						},
					},
				},
			},
		},
		entities.Resource{
			Arn:       "arn:aws:s3:::bucket3",
			AccountId: "11111",
		},
	).
	Build()

var InvalidTestUniverse_1 = entities.NewBuilder().
	WithPrincipals(
		entities.Principal{
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
	).
	WithAccounts(
		entities.Account{
			Id:    "88888",
			OrgId: "o-123",
			OrgPaths: []string{
				"o-123/",
				"o-123/ou-level-1/",
				"o-123/ou-level-1/ou-level-2/",
			},
			SCPs: [][]entities.Arn{
				{
					"arn:aws:organizations::00000:policy/o-aaa/service_control_policy/p-aaa/FullS3Access",
				},
			},
		},
	).
	Build()

var InvalidTestUniverse_2 = entities.NewBuilder().
	WithPrincipals(
		entities.Principal{
			Arn:       "arn:aws:iam::88888:role/role1",
			AccountId: "88888",
		},
	).
	WithResources(
		entities.Resource{
			Arn:       "arn:aws:s3:::bucket1",
			AccountId: "55555",
		},
	).
	WithAccounts(
		entities.Account{
			Id:    "55555",
			OrgId: "o-123",
			OrgPaths: []string{
				"o-123/",
				"o-123/ou-level-1/",
				"o-123/ou-level-1/ou-level-2/",
			},
			SCPs: [][]entities.Arn{
				{
					"arn:aws:organizations::00000:policy/o-aaa/service_control_policy/p-aaa/FullS3Access",
				},
			},
		},
	).
	Build()
