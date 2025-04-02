package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/entities"
)

// TestAuthContextKeys validates correct retrieval of Condition keys
func TestAuthContextKeys(t *testing.T) {
	type input struct {
		ac  AuthContext
		key string
	}

	tests := []testlib.TestCase[input, string]{
		{
			Name: "principal_tag",
			Input: input{
				ac: AuthContext{
					Principal: entities.Principal{
						Tags: []entities.Tag{
							{
								Key:   "baz",
								Value: "bam",
							},
							{
								Key:   "foo",
								Value: "bar",
							},
						},
					},
				},
				key: "aws:PrincipalTag/foo",
			},
			Want: "bar",
		},
		{
			Name: "principal_tag_bad_format",
			Input: input{
				ac: AuthContext{
					Principal: entities.Principal{
						Tags: []entities.Tag{
							{
								Key:   "foo",
								Value: "bar",
							},
						},
					},
				},
				key: "aws:PrincipalTag/foo/and/more",
			},
			Want: "",
		},
		{
			Name: "principal_tag_does_not_exist",
			Input: input{
				ac: AuthContext{
					Principal: entities.Principal{
						Tags: []entities.Tag{
							{
								Key:   "baz",
								Value: "bam",
							},
							{
								Key:   "foo",
								Value: "bar",
							},
						},
					},
				},
				key: "aws:PrincipalTag/DNE",
			},
			Want: "",
		},
		{
			Name: "resource_tag",
			Input: input{
				ac: AuthContext{
					Resource: entities.Resource{
						Tags: []entities.Tag{
							{
								Key:   "baz",
								Value: "bam",
							},
							{
								Key:   "foo",
								Value: "bar",
							},
						},
					},
				},
				key: "aws:ResourceTag/foo",
			},
			Want: "bar",
		},
		{
			Name: "request_tag",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:RequestTag/foo": "bar",
					},
				},
				key: "aws:RequestTag/foo",
			},
			Want: "bar",
		},
		{
			Name: "principal_service_check",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:PrincipalIsAWSService": "true",
					},
				},
				key: "aws:PrincipalIsAWSService",
			},
			Want: "true",
		},
		{
			Name: "principal_service_name",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:PrincipalServiceName": "cloudtrail.amazonaws.com",
					},
				},
				key: "aws:PrincipalServiceName",
			},
			Want: "cloudtrail.amazonaws.com",
		},
		{
			Name: "principal_type_role",
			Input: input{
				ac: AuthContext{
					Principal: entities.Principal{
						Type: "AWS::IAM::Role",
						Arn:  "arn:aws:iam::88888:role/somerole",
					},
				},
				key: "aws:PrincipalType",
			},
			Want: "Role", // TODO(nsiow) check casing/values
		},
		{
			Name: "principal_type_user",
			Input: input{
				ac: AuthContext{
					Principal: entities.Principal{
						Type: "AWS::IAM::User",
						Arn:  "arn:aws:iam::88888:user/someuser",
					},
				},
				key: "aws:PrincipalType",
			},
			Want: "User", // TODO(nsiow) check casing/values
		},
		{
			Name: "principal_type_user",
			Input: input{
				ac: AuthContext{
					Principal: entities.Principal{
						Type: "AWS::IAM::SomeNewEntityType",
						Arn:  "arn:aws:iam::88888:thing/foo",
					},
				},
				key: "aws:PrincipalType",
			},
			Want: "",
		},
		{
			Name: "principal_account",
			Input: input{
				ac: AuthContext{
					Principal: entities.Principal{
						AccountId: "55555",
					},
				},
				key: "aws:PrincipalAccount",
			},
			Want: "55555",
		},
		{
			Name: "principal_org_id",
			Input: input{
				ac: AuthContext{
					Principal: entities.Principal{
						Account: entities.Account{
							OrgId: "o-123",
						},
					},
				},
				key: "aws:PrincipalOrgID",
			},
			Want: "o-123",
		},
		{
			Name: "resource_account",
			Input: input{
				ac: AuthContext{
					Resource: entities.Resource{
						AccountId: "77777",
					},
				},
				key: "aws:ResourceAccount",
			},
			Want: "77777",
		},
		{
			Name: "resource_org_id",
			Input: input{
				ac: AuthContext{
					Resource: entities.Resource{
						Account: entities.Account{
							OrgId: "o-123",
						},
					},
				},
				key: "aws:ResourceOrgID",
			},
			Want: "o-123",
		},
		{
			Name: "source_arn",
			Input: input{
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SourceArn": "arn:aws:s3:::foo",
					},
				},
				key: "aws:SourceArn",
			},
			Want: "arn:aws:s3:::foo",
		},
		{
			Name: "current_time",
			Input: input{
				ac: AuthContext{
					Time: testlib.TestTime(),
				},
				key: "aws:CurrentTime",
			},
			Want: "2006-01-02T15:04:05",
		},
		{
			Name: "current_time_epoch",
			Input: input{
				ac: AuthContext{
					Time: testlib.TestTime(),
				},
				key: "aws:EpochTime",
			},
			Want: "1136214245",
		},
		{
			Name: "does_not_exist",
			Input: input{
				ac:  AuthContext{},
				key: "aws:ThisDoesNotExist",
			},
			Want: "",
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (string, error) {
		got := i.ac.Key(i.key)
		return got, nil
	})
}

// TestAuthContextMultiKeys validates correct retrieval of multi-valued Condition keys
func TestAuthContextMultiKeys(t *testing.T) {
	type input struct {
		ac  AuthContext
		key string
	}

	tests := []testlib.TestCase[input, []string]{
		{
			Name: "principal_tag",
			Input: input{
				ac: AuthContext{
					MultiValueProperties: map[string][]string{
						"aws:TagKeys": {
							"foo",
							"bar",
							"baz",
						},
					},
				},
				key: "aws:TagKeys",
			},
			Want: []string{"foo", "bar", "baz"},
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) ([]string, error) {
		got := i.ac.MultiKey(i.key)
		return got, nil
	})
}

// TestResolve validates the functionality of our variable resolution/expansion logic
func TestResolve(t *testing.T) {
	type input struct {
		str string
		ac  AuthContext
	}

	tests := []testlib.TestCase[input, string]{
		{
			Name: "simple_string",
			Input: input{
				str: "${aws:SomeStringKey}",
				ac: AuthContext{
					Properties: map[string]string{
						"aws:SomeStringKey": "SomeStringValue",
					},
				},
			},
			Want: "SomeStringValue",
		},
		{
			Name: "principal_tag",
			Input: input{
				str: "arn:aws:s3:::somebucket/${aws:PrincipalTag/foo}",
				ac: AuthContext{
					Principal: entities.Principal{
						Tags: []entities.Tag{
							{
								Key:   "foo",
								Value: "bar",
							},
						},
					},
				},
			},
			Want: "arn:aws:s3:::somebucket/bar",
		},
		{
			Name: "principal_tag_multi",
			Input: input{
				str: "arn:aws:s3:::somebucket/${aws:PrincipalTag/foo}/${aws:PrincipalTag/hello}",
				ac: AuthContext{
					Principal: entities.Principal{
						Tags: []entities.Tag{
							{
								Key:   "foo",
								Value: "bar",
							},
							{
								Key:   "hello",
								Value: "world",
							},
						},
					},
				},
			},
			Want: "arn:aws:s3:::somebucket/bar/world",
		},
		{
			Name: "absent_key",
			Input: input{
				str: "${aws:SomeStringKey}",
				ac:  AuthContext{},
			},
			Want: "",
		},
		{
			Name: "invalid_key",
			Input: input{
				str: "${aws:SomeStringKey}",
				ac:  AuthContext{},
			},
			Want: "",
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (string, error) {
		got := i.ac.Resolve(i.str)
		return got, nil
	})
}

// TestAuthContextReferenceTime is a minor test which validates that an unset time will result
// in a reference time of "now", which is measured by being "some time in the future"
func TestAuthContextReferenceTime(t *testing.T) {
	ac := AuthContext{}
	authContextTime := ac.now()

	if !authContextTime.After(testlib.TestTime()) {
		t.Fatalf("expected default time to be after our test reference time")
	}
}
