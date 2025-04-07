package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
)

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
					Principal: &entities.Principal{
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
					Principal: &entities.Principal{
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
					Principal: &entities.Principal{
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
					Resource: &entities.Resource{
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
					Properties: NewBagFromMap(map[string]string{
						"aws:RequestTag/foo": "bar",
					}),
				},
				key: "aws:RequestTag/foo",
			},
			Want: "bar",
		},
		{
			Name: "principal_service_check",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:PrincipalIsAWSService": "true",
					}),
				},
				key: "aws:PrincipalIsAWSService",
			},
			Want: "true",
		},
		{
			Name: "principal_service_name",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:PrincipalServiceName": "cloudtrail.amazonaws.com",
					}),
				},
				key: "aws:PrincipalServiceName",
			},
			Want: "cloudtrail.amazonaws.com",
		},
		{
			Name: "principal_type_role",
			Input: input{
				ac: AuthContext{
					Principal: &entities.Principal{
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
					Principal: &entities.Principal{
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
					Principal: &entities.Principal{
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
					Principal: &entities.Principal{
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
					Principal: &entities.Principal{
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
					Resource: &entities.Resource{
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
					Resource: &entities.Resource{
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceArn": "arn:aws:s3:::foo",
					}),
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
		got := i.ac.ConditionKey(i.key, TestingSimulationOptions)
		return got, nil
	})
}

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
					MultiValueProperties: NewBagFromMap(map[string][]string{
						"aws:TagKeys": {
							"foo",
							"bar",
							"baz",
						},
					}),
				},
				key: "aws:TagKeys",
			},
			Want: []string{"foo", "bar", "baz"},
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) ([]string, error) {
		got := i.ac.MultiKey(i.key, TestingSimulationOptions)
		return got, nil
	})
}

func TestSubstitute(t *testing.T) {
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
					Properties: NewBagFromMap(map[string]string{
						"aws:SomeStringKey": "SomeStringValue",
					}),
				},
			},
			Want: "SomeStringValue",
		},
		{
			Name: "principal_tag",
			Input: input{
				str: "arn:aws:s3:::somebucket/${aws:PrincipalTag/foo}",
				ac: AuthContext{
					Principal: &entities.Principal{
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
					Principal: &entities.Principal{
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
		got := i.ac.Substitute(i.str, TestingSimulationOptions)
		return got, nil
	})
}

func TestValidate(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, any]{
		{
			Name: "valid_auth_context",
			Input: AuthContext{
				Principal: &entities.Principal{},
				Action:    sar.MustLookupString("sqs:listqueues"),
			},
			ShouldErr: false,
		},
		{
			Name:      "empty_auth_context",
			Input:     AuthContext{},
			ShouldErr: true,
		},
		{
			Name: "missing_principal",
			Input: AuthContext{
				Resource: &entities.Resource{},
			},
			ShouldErr: true,
		},
		{
			Name: "missing_action",
			Input: AuthContext{
				Principal: &entities.Principal{},
			},
			ShouldErr: true,
		},
		{
			Name: "resource_unexpectedly_provided",
			Input: AuthContext{
				Principal: &entities.Principal{},
				Action:    sar.MustLookupString("sqs:listqueues"),
				Resource:  &entities.Resource{},
			},
			ShouldErr: true,
		},
		{
			Name: "wrong_resource_provided",
			Input: AuthContext{
				Principal: &entities.Principal{},
				Action:    sar.MustLookupString("sqs:getqueueurl"),
				Resource: &entities.Resource{
					Arn: "arn:aws:s3:::nsiow-test",
				},
			},
			ShouldErr: true,
		},
		{
			Name: "resource_unexpectedly_missing",
			Input: AuthContext{
				Principal: &entities.Principal{},
				Action:    sar.MustLookupString("sqs:getqueueurl"),
			},
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i AuthContext) (any, error) {
		return nil, i.Validate()
	})
}

func TestAuthContextReferenceTime(t *testing.T) {
	ac := AuthContext{}
	authContextTime := ac.now()

	if !authContextTime.After(testlib.TestTime()) {
		t.Fatalf("expected default time to be after our test reference time")
	}
}

func TestSARValidation(t *testing.T) {
	type input struct {
		ac  AuthContext
		key string
	}

	// FIXME(nsiow) write more of these tests... like so many more

	tests := []testlib.TestCase[input, string]{
		{
			Name: "supported_resource_tag",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("s3:getobject"),
					Properties: NewBagFromMap(map[string]string{
						"s3:authtype": "REST-HEADER",
					}),
				},
				key: "s3:AuthType",
			},
			Want: "REST-HEADER",
		},
		{
			Name: "unsupported_request_tag",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:FooBar": "some-value-here",
					}),
				},
				key: "aws:FooBar",
			},
			Want: "",
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (string, error) {
		got := i.ac.ConditionKey(i.key, NewOptions(WithFailOnUnknownConditionOperator()))
		return got, nil
	})
}

func TestSARValidationMultiKey(t *testing.T) {
	type input struct {
		ac  AuthContext
		key string
	}

	tests := []testlib.TestCase[input, []string]{
		{
			Name: "supported_resource_tag",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:createqueue"),
					MultiValueProperties: NewBagFromMap(map[string][]string{
						"aws:TagKeys": {
							"color",
							"temperature",
							"department",
						},
					}),
				},
				key: "aws:tagkeys",
			},
			Want: []string{
				"color",
				"temperature",
				"department",
			},
		},
		{
			Name: "unsupported_request_tag",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					MultiValueProperties: NewBagFromMap(map[string][]string{
						"s3:TagKeys": {
							"color",
							"temperature",
							"department",
						},
					}),
				},
				key: "s3:tagkeys",
			},
			Want: nil,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) ([]string, error) {
		got := i.ac.MultiKey(i.key, NewOptions(WithFailOnUnknownConditionOperator()))
		return got, nil
	})
}

func TestExtractTag(t *testing.T) {
	type input struct {
		ac   AuthContext
		key  string
		tags []entities.Tag
	}

	tests := []testlib.TestCase[input, string]{
		{
			Name: "valid_global_condition_key",
			Input: input{
				ac:  AuthContext{},
				key: "aws:PrincipalTag/color",
				tags: []entities.Tag{
					{
						Key:   "temperature",
						Value: "5",
					},
					{
						Key:   "color",
						Value: "blue",
					},
				},
			},
			Want: "blue",
		},
		{
			Name: "invalid_tag_structure",
			Input: input{
				ac:  AuthContext{},
				key: "color",
				tags: []entities.Tag{
					{
						Key:   "color",
						Value: "blue",
					},
				},
			},
			Want: "",
		},
		{
			Name: "missing_tag",
			Input: input{
				ac:  AuthContext{},
				key: "color",
				tags: []entities.Tag{
					{
						Key:   "temperature",
						Value: "5",
					},
				},
			},
			Want: "",
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (string, error) {
		got := i.ac.extractTag(i.key, i.tags)
		return got, nil
	})
}

func TestSupportsKey(t *testing.T) {
	type input struct {
		ac  AuthContext
		key string
	}

	tests := []testlib.TestCase[input, bool]{
		{
			Name: "valid_api_condition_key",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("s3:getobject"),
				},
				key: "s3:authtype",
			},
			Want: true,
		},
		{
			Name: "valid_resource_condition_key",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("dynamodb:query"),
					Resource: &entities.Resource{
						Arn: "arn:aws:dynamodb:us-west-2:55555:table/MyTable",
					},
				},
				key: "aws:resourcetag/foo",
			},
			Want: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		got := i.ac.supportsKey(i.key)
		return got, nil
	})
}
