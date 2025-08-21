package sim

import (
	"reflect"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

func TestExpandResources(t *testing.T) {
	sim, err := NewSimulator()
	if err != nil {
		t.Fatalf("error creating simulator: %v", err)
	}
	sim.Universe = SimpleTestUniverse_1

	expanded, err := sim.expandResources([]string{"arn:aws:s3:::bucket1"}, DEFAULT_OPTIONS)
	if err != nil {
		t.Fatalf("error expanding resources: %v", err)
	}

	expected := []string{"arn:aws:s3:::bucket1", "arn:aws:s3:::bucket1/*"}
	if !reflect.DeepEqual(expanded, expected) {
		t.Fatalf("expected %v but got: %v", expected, expanded)
	}

	_, err = sim.expandResources([]string{"arn:aws:s3:::404"}, DEFAULT_OPTIONS)
	if err == nil {
		t.Fatal("should have errored for missing bucket, but did not")
	}

	sim2, err := NewSimulator()
	if err != nil {
		t.Fatalf("error creating second simulator: %v", err)
	}
	sim2.Universe = entities.NewUniverse()

	sim2.Universe.PutResource(entities.Resource{
		Type: "AWS::S3::NotBucket",
		Arn:  "arn:aws:s3:::notabucket",
	})
	_, err = sim2.expandResources([]string{"arn:aws:s3:::notabucket"}, DEFAULT_OPTIONS)
	if err == nil {
		t.Fatal("should have errored for weird bucket, but did not")
	}
}

func TestNewSimulator(t *testing.T) {
	// Try with no options
	sim, err := NewSimulator()
	if err != nil {
		t.Fatalf("unexpected error creating a simulator with no options: %v", err)
	}
	if sim == nil {
		t.Fatalf("unexpected nil simulator when creating with no options")
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
		sim.Universe = i.uv
		res, err := sim.SimulateByArnWithOptions(
			i.principalArn,
			i.action,
			i.resourceArn,
			TestingSimulationOptions,
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
		opts    *Options
		actions []string
	}

	tests := []testlib.TestCase[input, map[string]int]{
		{
			Name: "simple_uv_1",
			Input: input{
				uv:      SimpleTestUniverse_1,
				actions: []string{"s3:listbucket"},
			},
			Want: map[string]int{
				"arn:aws:s3:::bucket1":   1,
				"arn:aws:s3:::bucket1/*": 1,
				"arn:aws:s3:::bucket2":   1,
				"arn:aws:s3:::bucket3":   0,
			},
		},
		{
			Name: "unrelated_actions",
			Input: input{
				uv:      SimpleTestUniverse_1,
				actions: []string{"sns:publish"},
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
			Name: "invalid_action",
			Input: input{
				uv:      entities.NewUniverse(),
				actions: []string{"foo:bar"},
			},
			ShouldErr: true,
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
		{
			Name: "cannot_expand_resources",
			Input: input{
				uv: InvalidTestUniverse_3,
			},
			ShouldErr: true,
		},
		{
			Name: "force_failure",
			Input: input{
				uv:      SimpleTestUniverse_1,
				actions: []string{"s3:listbucket"},
				opts:    &Options{ForceFailure: true},
			},
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (map[string]int, error) {
		if i.opts == nil {
			i.opts = &TestingSimulationOptions
		}

		sim, _ := NewSimulator()
		sim.Universe = i.uv
		summary, err := sim.AccessSummary(i.actions, *i.opts)
		if err != nil {
			return nil, err
		}

		return summary, nil
	})
}

func TestWhichPrincipals(t *testing.T) {
	type input struct {
		uv       *entities.Universe
		action   string
		resource string
		opts     *Options
	}

	tests := []testlib.TestCase[input, []string]{
		{
			Name: "simple_uv_1",
			Input: input{
				uv:       SimpleTestUniverse_1,
				action:   "s3:getobject",
				resource: "arn:aws:s3:::bucket1/object.txt",
			},
			Want: []string{
				"arn:aws:iam::88888:role/role1",
			},
		},
		{
			Name: "forced_failure",
			Input: input{
				uv:       SimpleTestUniverse_1,
				action:   "s3:getobject",
				resource: "arn:aws:s3:::bucket1/object.txt",
				opts:     &Options{ForceFailure: true},
			},
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) ([]string, error) {
		if i.opts == nil {
			i.opts = &TestingSimulationOptions
		}

		sim, _ := NewSimulator()
		sim.Universe = i.uv
		results, err := sim.WhichPrincipals(i.action, i.resource, *i.opts)
		if err != nil {
			return nil, err
		}

		return results, nil
	})
}

func TestWhichActions(t *testing.T) {
	type input struct {
		uv        *entities.Universe
		principal string
		resource  string
		opts      *Options
	}

	tests := []testlib.TestCase[input, []string]{
		{
			Name: "simple_uv_1",
			Input: input{
				uv:        SimpleTestUniverse_1,
				principal: "arn:aws:iam::88888:role/role1",
				resource:  "arn:aws:s3:::bucket1/object.txt",
			},
			Want: []string{
				"s3:GetObject",
				"s3:ListBucket",
			},
		},
		{
			Name: "forced_failure",
			Input: input{
				uv:        SimpleTestUniverse_1,
				principal: "arn:aws:iam::88888:role/role1",
				resource:  "arn:aws:s3:::bucket1/object.txt",
				opts:      &Options{ForceFailure: true},
			},
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) ([]string, error) {
		if i.opts == nil {
			i.opts = &TestingSimulationOptions
		}

		sim, _ := NewSimulator()
		sim.Universe = i.uv
		results, err := sim.WhichActions(i.principal, i.resource, *i.opts)
		if err != nil {
			return nil, err
		}

		return results, nil
	})
}

func TestWhichResources(t *testing.T) {
	type input struct {
		uv        *entities.Universe
		principal string
		action    string
		opts      *Options
	}

	tests := []testlib.TestCase[input, []string]{
		{
			Name: "simple_uv_1",
			Input: input{
				uv:        SimpleTestUniverse_1,
				principal: "arn:aws:iam::88888:role/role1",
				action:    "s3:getobject",
			},
			Want: []string{
				"arn:aws:s3:::bucket1/*",
			},
		},
		{
			Name: "forced_failure",
			Input: input{
				uv:        SimpleTestUniverse_1,
				principal: "arn:aws:iam::88888:role/role1",
				action:    "s3:getobject",
				opts:      &Options{DefaultS3Key: "*", ForceFailure: true},
			},
			ShouldErr: true,
		},
		{
			Name: "cannot_expand_resources",
			Input: input{
				uv:        InvalidTestUniverse_3,
				principal: "arn:aws:iam::88888:role/role1",
				action:    "s3:getobject",
			},
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) ([]string, error) {
		if i.opts == nil {
			i.opts = &TestingSimulationOptions
		}

		sim, _ := NewSimulator()
		sim.Universe = i.uv
		results, err := sim.WhichResources(i.principal, i.action, *i.opts)
		if err != nil {
			return nil, err
		}

		return results, nil
	})
}

var SimpleTestUniverse_1 = entities.NewBuilder().
	WithPrincipals(
		entities.Principal{
			Arn:       "arn:aws:iam::88888:role/role1",
			Type:      "AWS::IAM::Role",
			AccountId: "88888",
			InlinePolicies: []policy.Policy{
				{
					Statement: []policy.Statement{
						{
							Effect: policy.EFFECT_ALLOW,
							Action: []string{
								"s3:listbucket",
								"s3:getobject",
							},
						},
						{
							Effect:   policy.EFFECT_ALLOW,
							Action:   []string{"s3:get"},
							Resource: []string{"*"},
						},
					},
				},
			},
		},
		entities.Principal{
			Arn:       "arn:aws:iam::88888:role/role2",
			Type:      "AWS::IAM::Role",
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
			Type:      "AWS::IAM::Role",
			AccountId: "11111",
		},
	).
	WithResources(
		entities.Resource{
			Arn:       "arn:aws:s3:::bucket1",
			Type:      "AWS::S3::Bucket",
			AccountId: "88888",
		},
		entities.Resource{
			Arn:       "arn:aws:s3:::bucket2",
			Type:      "AWS::S3::Bucket",
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
			Type:      "AWS::S3::Bucket",
			AccountId: "11111",
		},
	).
	Build()

var InvalidTestUniverse_1 = entities.NewBuilder().
	WithPrincipals(
		entities.Principal{
			Arn:       "arn:aws:iam::88888:role/role1",
			Type:      "AWS::IAM::Role",
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
			OrgNodes: []entities.OrgNode{
				{
					SCPs: []entities.Arn{
						"arn:aws:organizations::00000:policy/o-aaa/service_control_policy/p-aaa/FullS3Access",
					},
				},
			},
		},
	).
	Build()

var InvalidTestUniverse_2 = entities.NewBuilder().
	WithPrincipals(
		entities.Principal{
			Arn:       "arn:aws:iam::88888:role/role1",
			Type:      "AWS::IAM::Role",
			AccountId: "88888",
		},
	).
	WithResources(
		entities.Resource{
			Arn:       "arn:aws:s3:::bucket1",
			Type:      "AWS::S3::Bucket",
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
			OrgNodes: []entities.OrgNode{
				{
					SCPs: []entities.Arn{
						"arn:aws:organizations::00000:policy/o-aaa/service_control_policy/p-aaa/FullS3Access",
					},
				},
			},
		},
	).
	Build()

var InvalidTestUniverse_3 = entities.NewBuilder().
	WithResources(
		entities.Resource{
			Arn:       "arn:aws:s3:::notabucket",
			Type:      "AWS::S3::NotABucket",
			AccountId: "55555",
		},
	).
	Build()
