package awsconfig

import (
	"os"
	"path"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

func TestLoad(t *testing.T) {
	tests := []testlib.TestCase[string, entities.Universe]{

		// ---------------------------------------------------------------------------------------------
		// Valid
		// ---------------------------------------------------------------------------------------------

		{
			Name:  "base_valid_empty",
			Input: `../../../testdata/config_loading/base_valid_empty.json`,
			Want:  *entities.NewUniverse(),
		},
		{
			Name:  "base_valid_empty_json_l",
			Input: `../../../testdata/config_loading/base_valid_empty.jsonl`,
			Want:  *entities.NewUniverse(),
		},
		{
			Name:  "account_valid",
			Input: `../../../testdata/config_loading/account_valid.json`,
			Want: *entities.NewBuilder().
				WithAccounts(
					entities.Account{
						Id:    "000000000000",
						OrgId: "o-123",
						OrgPaths: []string{
							"o-123/",
							"o-123/ou-level-1/",
							"o-123/ou-level-1/ou-level-2/",
						},
						SCPs: [][]entities.Arn{
							{
								"arn:aws:organizations::aws:policy/service_control_policy/p-FullAWSAccess/FullAWSAccess",
							},
							{
								"arn:aws:organizations::000000000000:policy/o-aaa/service_control_policy/p-aaa/FullS3Access",
							},
						},
					},
				).
				Build(),
		},
		{
			Name:  "scp_valid",
			Input: `../../../testdata/config_loading/scp_valid.json`,
			Want: *entities.NewBuilder().
				WithPolicies(
					entities.ManagedPolicy{
						Type:      "Yams::Organizations::ServiceControlPolicy",
						AccountId: "000000000000",
						Arn:       "arn:aws:organizations::000000000000:policy/o-aaa/service_control_policy/p-aaa/FullS3Access",
						Policy: policy.Policy{
							Version: "2012-10-17",
							Id:      "",
							Statement: policy.StatementBlock{
								policy.Statement{
									Sid:    "",
									Effect: "Allow",
									Action: policy.Value{
										"s3:*",
									},
									Resource: policy.Value{
										"*",
									},
								},
							},
						},
					},
				).
				Build(),
		},
		{
			Name:  "group_valid",
			Input: `../../../testdata/config_loading/group_valid.json`,
			Want: *entities.NewBuilder().
				WithGroups(
					entities.Group{
						Type:      "AWS::IAM::Group",
						AccountId: "000000000000",
						Arn:       "arn:aws:iam::000000000000:group/family",
						InlinePolicies: []policy.Policy{
							{
								Version: "2012-10-17",
								Id:      "",
								Statement: policy.StatementBlock{
									policy.Statement{
										Sid:    "",
										Effect: "Allow",
										Action: policy.Value{
											"s3:*",
										},
										Resource: policy.Value{
											"*",
										},
									},
								},
							},
						},
					},
				).
				Build(),
		},

		// ---------------------------------------------------------------------------------------------
		// Invalid
		// ---------------------------------------------------------------------------------------------

		{
			Name:      "base_invalid",
			Input:     `../../../testdata/config_loading/base_invalid.json`,
			ShouldErr: true,
		},
		{
			Name:      "base_invalid_jsonl",
			Input:     `../../../testdata/config_loading/base_invalid.jsonl`,
			ShouldErr: true,
		},
		{
			Name:      "account_invalid_scp",
			Input:     `../../../testdata/config_loading/account_invalid_scp.json`,
			ShouldErr: true,
		},
		{
			Name:      "scp_invalid_syntax",
			Input:     `../../../testdata/config_loading/scp_invalid_syntax.json`,
			ShouldErr: true,
		},
		{
			Name:      "scp_invalid_syntax_2",
			Input:     `../../../testdata/config_loading/scp_invalid_syntax_2.json`,
			ShouldErr: true,
		},
		{
			Name:      "group_invalid_bad_shape",
			Input:     `../../../testdata/config_loading/group_invalid_bad_shape.json`,
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(fp string) (entities.Universe, error) {
		// Load test data
		f, err := os.Open(fp)
		if err != nil {
			t.Fatalf("error while attempting to open test file '%s': %v", fp, err)
		}

		// Call the correct loader based on input type
		l := NewLoader()
		ext := path.Ext(fp)
		switch ext {
		case ".json":
			err = l.LoadJson(f)
		case ".jsonl":
			err = l.LoadJsonl(f)
		default:
			t.Fatalf("unsure how to handle ext '%s'", ext)
		}

		// Handle loading errors; these may be expected
		if err != nil {
			return entities.Universe{}, err
		}
		return *l.Universe(), nil
	})
}

func TestLoad_EdgeCases(t *testing.T) {
	reader := &testlib.FailReader{}
	l := NewLoader()

	err := l.LoadJson(reader)
	if err == nil {
		t.Fatalf("LoadJson should have failed, but succeeded")
	}

	err = l.LoadJsonl(reader)
	if err == nil {
		t.Fatalf("LoadJson; should have failed, but succeeded")
	}
}
