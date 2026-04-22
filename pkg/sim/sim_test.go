package sim

import (
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

func TestNewPlaceholderResource(t *testing.T) {
	tests := []struct {
		name     string
		arn      string
		wantAcct string
	}{
		{"explicit_account", "arn:aws:iam::88888:role/MyRole", "88888"},
		{"wildcard_account", "arn:aws:iam::*:role/MyRole", "*"},
		{"missing_account_s3", "arn:aws:s3:::mybucket", "*"},
		{"short_arn", "*", "*"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newPlaceholderResource(tc.arn)
			if r.AccountId != tc.wantAcct {
				t.Fatalf("AccountId: want %q, got %q", tc.wantAcct, r.AccountId)
			}
			if r.Arn != tc.arn {
				t.Fatalf("Arn: want %q, got %q", tc.arn, r.Arn)
			}
		})
	}
}

func TestSpecializeForPrincipal(t *testing.T) {
	p := &entities.FrozenPrincipal{AccountId: "88888"}

	// Wildcard account: substituted with principal's account.
	wild := &entities.FrozenResource{
		AccountId:   "*",
		Arn:         "arn:aws:iam::*:role/MyRole",
		ArnSegments: entities.SplitArn("arn:aws:iam::*:role/MyRole"),
	}
	got := specializeForPrincipal(wild, p)
	if got.AccountId != "88888" {
		t.Fatalf("wildcard: AccountId want 88888, got %q", got.AccountId)
	}
	if got.Arn != "arn:aws:iam::88888:role/MyRole" {
		t.Fatalf("wildcard: Arn want arn:aws:iam::88888:role/MyRole, got %q", got.Arn)
	}

	// Explicit account: returned unchanged.
	explicit := &entities.FrozenResource{
		AccountId:   "11111",
		Arn:         "arn:aws:iam::11111:role/MyRole",
		ArnSegments: entities.SplitArn("arn:aws:iam::11111:role/MyRole"),
	}
	if specializeForPrincipal(explicit, p) != explicit {
		t.Fatal("explicit: expected same pointer returned")
	}

	// Nil inputs: returned unchanged.
	if specializeForPrincipal(nil, p) != nil {
		t.Fatal("nil resource: expected nil")
	}
	if specializeForPrincipal(wild, nil) != wild {
		t.Fatal("nil principal: expected same pointer returned")
	}
}

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
				"arn:aws:s3:::bucket1": 1,
				"arn:aws:s3:::bucket2": 1,
				"arn:aws:s3:::bucket3": 0,
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
			Name: "create_action_nonexistent_resource",
			Input: input{
				uv:       CreateActionTestUniverse,
				action:   "sqs:createqueue",
				resource: "arn:aws:sqs:us-east-1:88888:newqueue",
			},
			Want: []string{
				"arn:aws:iam::88888:role/role1",
			},
		},
		{
			Name: "non_create_action_nonexistent_resource",
			Input: input{
				uv:       CreateActionTestUniverse,
				action:   "sqs:sendmessage",
				resource: "arn:aws:sqs:us-east-1:88888:nonexistent",
			},
			ShouldErr: true,
		},
		{
			// An explicit account in the placeholder ARN scopes the Create to that account;
			// only principals in 88888 are returned.
			Name: "create_explicit_account_scopes_to_that_account",
			Input: input{
				uv:       CrossAccountCreateTestUniverse,
				action:   "iam:createrole",
				resource: "arn:aws:iam::88888:role/NewRole",
			},
			Want: []string{
				"arn:aws:iam::88888:role/creator",
			},
		},
		{
			// A wildcard in the account segment means "any account" — the placeholder's account is
			// substituted with each principal's account at sim time, so every permitted principal
			// passes the same-account check.
			Name: "create_wildcard_account_returns_all_permitted",
			Input: input{
				uv:       CrossAccountCreateTestUniverse,
				action:   "iam:createrole",
				resource: "arn:aws:iam::*:role/NewRole",
			},
			Want: []string{
				"arn:aws:iam::11111:role/creator",
				"arn:aws:iam::88888:role/creator",
			},
		},
		{
			// S3 bucket ARNs have no account segment; treat same as wildcard.
			Name: "create_empty_account_s3_bucket_returns_all_permitted",
			Input: input{
				uv:       CrossAccountCreateTestUniverse,
				action:   "s3:createbucket",
				resource: "arn:aws:s3:::any-bucket",
			},
			Want: []string{
				"arn:aws:iam::11111:role/creator",
				"arn:aws:iam::88888:role/creator",
			},
		},
		{
			// RunInstances uses the same Create heuristic; explicit account scopes same-account.
			Name: "runinstances_explicit_account_scopes_to_that_account",
			Input: input{
				uv:       CrossAccountCreateTestUniverse,
				action:   "ec2:runinstances",
				resource: "arn:aws:ec2:us-east-1:11111:instance/i-new",
			},
			Want: []string{
				"arn:aws:iam::11111:role/creator",
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

		// sort for deterministic comparison; goroutine scheduling can shuffle order
		sort.Strings(results)
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
				// s3:ListBucket doesn't target objects, only buckets
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

var CreateActionTestUniverse = entities.NewBuilder().
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
							Action:   []string{"sqs:createqueue", "sqs:sendmessage"},
							Resource: []string{"*"},
						},
					},
				},
			},
		},
	).
	Build()

// CrossAccountCreateTestUniverse has principals in two accounts (88888 and 11111) each with
// iam:CreateRole + s3:CreateBucket + ec2:RunInstances on *. Used to verify that Create
// placeholders with an explicit account scope results to that account, while wildcard /
// account-less ARNs return all permitted principals.
var CrossAccountCreateTestUniverse = entities.NewBuilder().
	WithPrincipals(
		entities.Principal{
			Arn:       "arn:aws:iam::88888:role/creator",
			Type:      "AWS::IAM::Role",
			AccountId: "88888",
			InlinePolicies: []policy.Policy{
				{
					Statement: []policy.Statement{
						{
							Effect: policy.EFFECT_ALLOW,
							Action: []string{
								"iam:createrole",
								"s3:createbucket",
								"ec2:runinstances",
							},
							Resource: []string{"*"},
						},
					},
				},
			},
		},
		entities.Principal{
			Arn:       "arn:aws:iam::11111:role/creator",
			Type:      "AWS::IAM::Role",
			AccountId: "11111",
			InlinePolicies: []policy.Policy{
				{
					Statement: []policy.Statement{
						{
							Effect: policy.EFFECT_ALLOW,
							Action: []string{
								"iam:createrole",
								"s3:createbucket",
								"ec2:runinstances",
							},
							Resource: []string{"*"},
						},
					},
				},
			},
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
					SCPs: []entities.OrgPolicyRef{
						{Arn: "arn:aws:organizations::00000:policy/o-aaa/service_control_policy/p-aaa/FullS3Access"},
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
					SCPs: []entities.OrgPolicyRef{
						{Arn: "arn:aws:organizations::00000:policy/o-aaa/service_control_policy/p-aaa/FullS3Access"},
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

func TestSimulateByArn_FuzzyMatch(t *testing.T) {
	type input struct {
		uv           *entities.Universe
		action       string
		principalArn string
		resourceArn  string
		opts         Options
	}

	tests := []testlib.TestCase[input, bool]{
		{
			Name: "fuzzy_match_principal",
			Input: input{
				uv:           SimpleTestUniverse_1,
				action:       "s3:listbucket",
				principalArn: "role1", // partial match
				resourceArn:  "arn:aws:s3:::bucket1",
				opts:         Options{EnableFuzzyMatchArn: true, DefaultS3Key: "*"},
			},
			Want: true,
		},
		{
			Name: "fuzzy_match_resource",
			Input: input{
				uv:           SimpleTestUniverse_1,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/role1",
				resourceArn:  "bucket1", // partial match
				opts:         Options{EnableFuzzyMatchArn: true, DefaultS3Key: "*"},
			},
			Want: true,
		},
		{
			Name: "fuzzy_match_no_match",
			Input: input{
				uv:           SimpleTestUniverse_1,
				action:       "s3:listbucket",
				principalArn: "nonexistent",
				resourceArn:  "arn:aws:s3:::bucket1",
				opts:         Options{EnableFuzzyMatchArn: true, DefaultS3Key: "*"},
			},
			ShouldErr: true,
		},
		{
			Name: "fuzzy_match_too_many_principal_matches",
			Input: input{
				uv:           SimpleTestUniverse_1,
				action:       "s3:listbucket",
				principalArn: "role", // matches role1, role2, role3
				resourceArn:  "arn:aws:s3:::bucket1",
				opts:         Options{EnableFuzzyMatchArn: true, DefaultS3Key: "*"},
			},
			ShouldErr: true,
		},
		{
			Name: "fuzzy_match_too_many_resource_matches",
			Input: input{
				uv:           SimpleTestUniverse_1,
				action:       "s3:listbucket",
				principalArn: "arn:aws:iam::88888:role/role1",
				resourceArn:  "bucket", // matches bucket1, bucket2, bucket3
				opts:         Options{EnableFuzzyMatchArn: true, DefaultS3Key: "*"},
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
			i.opts,
		)
		if err != nil {
			return false, err
		}

		return res.IsAllowed, nil
	})
}

func TestSimulateByArn_CreateAction(t *testing.T) {
	// Test case where action is Create* and resource doesn't exist
	sim, _ := NewSimulator()
	sim.Universe = SimpleTestUniverse_1

	// s3:CreateBucket is a Create action without resources in SAR, so let's use something else
	// Actually test the CreateBucket path which checks for !ok && strings.HasPrefix(ac.Action.Name, "Create")
	uv := entities.NewBuilder().
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
								Action:   []string{"sqs:createqueue"},
								Resource: []string{"*"},
							},
						},
					},
				},
			},
		).
		Build()

	sim.Universe = uv
	// CreateQueue with non-existent resource ARN
	res, err := sim.SimulateByArnWithOptions(
		"arn:aws:iam::88888:role/role1",
		"sqs:createqueue",
		"arn:aws:sqs:us-east-1:88888:newqueue",
		Options{DefaultS3Key: "*"},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.IsAllowed {
		t.Fatal("expected allow for create action with non-existent resource")
	}
}

// TestSimulateByArn_CreateAction_CrossAccount verifies a principal cannot create a resource
// in a different account even when its identity policy permits the action.
func TestSimulateByArn_CreateAction_CrossAccount(t *testing.T) {
	uv := entities.NewBuilder().
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
								Action:   []string{"iam:createrole"},
								Resource: []string{"*"},
							},
						},
					},
				},
			},
		).
		Build()

	sim, _ := NewSimulator()
	sim.Universe = uv

	// Principal in 88888 attempting to create a role in 11111 — should deny.
	res, err := sim.SimulateByArnWithOptions(
		"arn:aws:iam::88888:role/role1",
		"iam:createrole",
		"arn:aws:iam::11111:role/NewRole",
		Options{},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsAllowed {
		t.Fatal("expected deny for x-account create action")
	}

	// Same principal, same-account target — should allow.
	res, err = sim.SimulateByArnWithOptions(
		"arn:aws:iam::88888:role/role1",
		"iam:createrole",
		"arn:aws:iam::88888:role/NewRole",
		Options{},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsAllowed {
		t.Fatal("expected allow for same-account create action")
	}

	// Wildcard account — substituted with principal's account, should allow.
	res, err = sim.SimulateByArnWithOptions(
		"arn:aws:iam::88888:role/role1",
		"iam:createrole",
		"arn:aws:iam::*:role/NewRole",
		Options{},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsAllowed {
		t.Fatal("expected allow for wildcard-account create action")
	}
}

func TestSimulateByArn_RunInstancesAction(t *testing.T) {
	// Test case where action is RunInstances and resource doesn't exist
	uv := entities.NewBuilder().
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
								Action:   []string{"ec2:RunInstances"},
								Resource: []string{"*"},
							},
						},
					},
				},
			},
		).
		Build()

	sim, _ := NewSimulator()
	sim.Universe = uv

	// RunInstances with non-existent resource ARN
	res, err := sim.SimulateByArnWithOptions(
		"arn:aws:iam::88888:role/role1",
		"ec2:RunInstances",
		"arn:aws:ec2:us-east-1:88888:instance/i-nonexistent",
		Options{DefaultS3Key: "*"},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.IsAllowed {
		t.Fatal("expected allow for RunInstances action with non-existent resource")
	}
}

func TestSimulateByArn_Default(t *testing.T) {
	// Test the default SimulateByArn function (which uses DEFAULT_OPTIONS)
	uv := entities.NewBuilder().
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
								Resource: []string{"arn:aws:s3:::mybucket"},
							},
						},
					},
				},
			},
		).
		WithResources(
			entities.Resource{
				Arn:       "arn:aws:s3:::mybucket",
				Type:      "AWS::S3::Bucket",
				AccountId: "88888",
			},
		).
		Build()

	sim, _ := NewSimulator()
	sim.Universe = uv

	// Use the default SimulateByArn (no options)
	res, err := sim.SimulateByArn(
		"arn:aws:iam::88888:role/role1",
		"s3:listbucket",
		"arn:aws:s3:::mybucket",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.IsAllowed {
		t.Fatal("expected allow")
	}
}

func TestSimulateWithOptions_ForceFailure(t *testing.T) {
	sim, _ := NewSimulator()
	sim.Universe = SimpleTestUniverse_1

	ac := AuthContext{
		Action: sar.MustLookupString("s3:listbucket"),
		Principal: &entities.FrozenPrincipal{
			Arn:       "arn:aws:iam::88888:role/role1",
			AccountId: "88888",
		},
		Resource: &entities.FrozenResource{
			Arn:       "arn:aws:s3:::bucket1",
			AccountId: "88888",
		},
	}

	opts := Options{ForceFailure: true}
	_, err := sim.SimulateWithOptions(ac, opts)

	if err == nil {
		t.Fatal("expected error due to ForceFailure option")
	}
}

func TestProduct_BatchSubmission(t *testing.T) {
	// Set a small batch size to trigger the mid-batch submission path
	os.Setenv("YAMS_SIM_BATCH_SIZE", "2")
	defer os.Unsetenv("YAMS_SIM_BATCH_SIZE")

	// Create a universe with enough entities to exceed batch size
	uv := entities.NewBuilder().
		WithPrincipals(
			entities.Principal{
				Arn:       "arn:aws:iam::88888:role/role1",
				Type:      "AWS::IAM::Role",
				AccountId: "88888",
			},
			entities.Principal{
				Arn:       "arn:aws:iam::88888:role/role2",
				Type:      "AWS::IAM::Role",
				AccountId: "88888",
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
				AccountId: "88888",
			},
		).
		Build()

	sim, _ := NewSimulator()
	sim.Universe = uv

	// This should trigger batch submission when jobs exceed batch size
	results, err := sim.AccessSummary([]string{"s3:listbucket"}, TestingSimulationOptions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify results were computed
	if len(results) == 0 {
		t.Fatal("expected results from AccessSummary")
	}
}

// Test public wrappers that delegate to unexported methods
func TestExpandResources_PublicWrapper(t *testing.T) {
	sim, err := NewSimulator()
	if err != nil {
		t.Fatalf("error creating simulator: %v", err)
	}
	sim.Universe = SimpleTestUniverse_1

	expanded, err := sim.ExpandResources([]string{"arn:aws:s3:::bucket1"}, DEFAULT_OPTIONS)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	expected := []string{"arn:aws:s3:::bucket1", "arn:aws:s3:::bucket1/*"}
	if !reflect.DeepEqual(expanded, expected) {
		t.Fatalf("expected %v but got: %v", expected, expanded)
	}
}

func TestFreezeResources_PublicWrapper(t *testing.T) {
	sim, err := NewSimulator()
	if err != nil {
		t.Fatalf("error creating simulator: %v", err)
	}
	sim.Universe = SimpleTestUniverse_1

	frozen, err := sim.FreezeResources([]string{"arn:aws:s3:::bucket1"}, DEFAULT_OPTIONS)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(frozen) == 0 {
		t.Fatal("expected frozen resources")
	}
}

func TestProductFrozenStreaming(t *testing.T) {
	sim, err := NewSimulator()
	if err != nil {
		t.Fatalf("error creating simulator: %v", err)
	}
	sim.Universe = SimpleTestUniverse_1

	pArns := []string{}
	for p := range sim.Universe.Principals() {
		pArns = append(pArns, p.Arn)
	}
	principals, err := sim.FreezePrincipals(pArns, TestingSimulationOptions)
	if err != nil {
		t.Fatalf("error freezing principals: %v", err)
	}

	resources, err := sim.FreezeResources(
		[]string{"arn:aws:s3:::bucket1"}, TestingSimulationOptions)
	if err != nil {
		t.Fatalf("error freezing resources: %v", err)
	}

	// Test happy path
	count := 0
	err = sim.ProductFrozenStreaming(principals, []string{"s3:listbucket"}, resources,
		TestingSimulationOptions, func(r AccessTuple) {
			count++
		})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count == 0 {
		t.Fatal("expected streaming results")
	}

	// Test unknown action
	err = sim.ProductFrozenStreaming(principals, []string{"fake:unknown"}, resources,
		TestingSimulationOptions, func(r AccessTuple) {})
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
}
