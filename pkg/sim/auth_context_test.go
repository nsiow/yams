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

		// -------------------------------------------------------------------------------------------
		// Global keys; default handling
		// -------------------------------------------------------------------------------------------

		{
			Name: "called_via_first",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:CalledViaFirst": "anchovy",
					}),
				},
				key: "aws:calledviafirst",
			},
			Want: "anchovy",
		},
		{
			Name: "called_via_last",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:CalledViaLast": "bass",
					}),
				},
				key: "aws:calledvialast",
			},
			Want: "bass",
		},
		{
			Name: "ec2_instance_source_private_ipv4",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:Ec2InstanceSourcePrivateIPv4": "carp",
					}),
				},
				key: "aws:ec2instancesourceprivateipv4",
			},
			Want: "carp",
		},
		{
			Name: "ec2_instance_source_vpc",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:Ec2InstanceSourceVPC": "eel",
					}),
				},
				key: "aws:ec2instancesourcevpc",
			},
			Want: "eel",
		},
		{
			Name: "federated_provider",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:FederatedProvider": "flounder",
					}),
				},
				key: "aws:federatedprovider",
			},
			Want: "flounder",
		},
		{
			Name: "multi_factor_auth_age",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:MultiFactorAuthAge": "guppy",
					}),
				},
				key: "aws:multifactorauthage",
			},
			Want: "guppy",
		},
		{
			Name: "multi_factor_auth_present",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:MultiFactorAuthPresent": "halibut",
					}),
				},
				key: "aws:multifactorauthpresent",
			},
			Want: "halibut",
		},
		{
			Name: "principal_service_names_list",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:PrincipalServiceNamesList": "herring",
					}),
				},
				key: "aws:principalservicenameslist",
			},
			Want: "herring",
		},

		{
			Name: "referer",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:Referer": "mackerel",
					}),
				},
				key: "aws:referer",
			},
			Want: "mackerel",
		},
		{
			Name: "requested_region",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:RequestedRegion": "salmon",
					}),
				},
				key: "aws:requestedregion",
			},
			Want: "salmon",
		},
		{
			Name: "role_delivery",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"ec2:RoleDelivery": "trout",
					}),
				},
				key: "ec2:roledelivery",
			},
			Want: "trout",
		},
		{
			Name: "secure_transport",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:SecureTransport": "tuna",
					}),
				},
				key: "aws:securetransport",
			},
			Want: "tuna",
		},
		{
			Name: "source_account",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceAccount": "urchin",
					}),
				},
				key: "aws:sourceaccount",
			},
			Want: "urchin",
		},
		{
			Name: "source_arn",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceArn": "walleye",
					}),
				},
				key: "aws:sourcearn",
			},
			Want: "walleye",
		},
		{
			Name: "source_identity",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIdentity": "xenopus",
					}),
				},
				key: "aws:sourceidentity",
			},
			Want: "xenopus",
		},
		{
			Name: "source_instance_arn",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"ec2:SourceInstanceArn": "yellowfin",
					}),
				},
				key: "ec2:sourceinstancearn",
			},
			Want: "yellowfin",
		},
		{
			Name: "source_ip",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIp": "zander",
					}),
				},
				key: "aws:sourceip",
			},
			Want: "zander",
		},
		{
			Name: "source_org_id",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceOrgID": "abalone",
					}),
				},
				key: "aws:sourceorgid",
			},
			Want: "abalone",
		},
		{
			Name: "source_vpc",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceVpc": "blenny",
					}),
				},
				key: "aws:sourcevpc",
			},
			Want: "blenny",
		},
		{
			Name: "source_vpce",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceVPCE": "cod",
					}),
				},
				key: "aws:sourcevpce",
			},
			Want: "cod",
		},
		{
			Name: "token_issue_time",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:TokenIssueTime": "dorado",
					}),
				},
				key: "aws:tokenissuetime",
			},
			Want: "dorado",
		},
		{
			Name: "user_agent",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:UserAgent": "escolar",
					}),
				},
				key: "aws:useragent",
			},
			Want: "escolar",
		},
		{
			Name: "user_id",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:UserId": "fluke",
					}),
				},
				key: "aws:userid",
			},
			Want: "fluke",
		},
		{
			Name: "username",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:Username": "gurnard",
					}),
				},
				key: "aws:username",
			},
			Want: "gurnard",
		},
		{
			Name: "via_aws_service",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:ViaAWSService": "haddock",
					}),
				},
				key: "aws:viaawsservice",
			},
			Want: "haddock",
		},
		{
			Name: "vpc_source_ip",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"aws:VpcSourceIP": "ilexander",
					}),
				},
				key: "aws:vpcsourceip",
			},
			Want: "ilexander",
		},

		// -------------------------------------------------------------------------------------------
		// Global keys; special handling
		// -------------------------------------------------------------------------------------------

		{
			Name: "principal_tag",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
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
					Principal: &entities.FrozenPrincipal{
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
			Want: EMPTY,
		},
		{
			Name: "principal_tag_does_not_exist",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
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
			Want: EMPTY,
		},
		{
			Name: "resource_tag_based_on_valid_resource",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:receivemessage"),
					Resource: &entities.FrozenResource{
						Arn: "arn:aws:sqs:us-east-1:55555:myqueue",
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
					Action: sar.MustLookupString("sqs:createqueue"),
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
				ac:  AuthContext{},
				key: "aws:PrincipalIsAWSService",
			},
			Want: "false", // we do not support AWS service evaluation
		},
		{
			Name: "principal_service_check_override",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:PrincipalIsAWSService": "true",
					}),
				},
				key: "aws:PrincipalIsAWSService",
			},
			Want: "true", // but do allow overrides
		},
		{
			Name: "principal_service_name",
			Input: input{
				ac:  AuthContext{},
				key: "aws:PrincipalServiceName",
			},
			Want: EMPTY, // we do not support AWS service evaluation
		},
		{
			Name: "principal_service_name_override",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:PrincipalServiceName": "cloudtrail.amazonaws.com",
					}),
				},
				key: "aws:PrincipalServiceName",
			},
			Want: "cloudtrail.amazonaws.com", // but do allow overrides
		},
		{
			Name: "principal_type_role",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						Type: "AWS::IAM::Role",
						Arn:  "arn:aws:iam::88888:role/somerole",
					},
				},
				key: "aws:PrincipalType",
			},
			Want: "AssumedRole",
		},
		{
			Name: "principal_type_user",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						Type: "AWS::IAM::User",
						Arn:  "arn:aws:iam::88888:user/someuser",
					},
				},
				key: "aws:PrincipalType",
			},
			Want: "User",
		},
		{
			Name: "principal_type_user",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
						Type: "AWS::IAM::SomeNewEntityType",
						Arn:  "arn:aws:iam::88888:thing/foo",
					},
				},
				key: "aws:PrincipalType",
			},
			Want: EMPTY,
		},
		{
			Name: "principal_account",
			Input: input{
				ac: AuthContext{
					Principal: &entities.FrozenPrincipal{
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
					Principal: &entities.FrozenPrincipal{
						Account: entities.FrozenAccount{
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
					Resource: &entities.FrozenResource{
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
					Resource: &entities.FrozenResource{
						Account: entities.FrozenAccount{
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
			Want: EMPTY,
		},

		// -------------------------------------------------------------------------------------------
		// Nil Principal/Resource edge cases
		// -------------------------------------------------------------------------------------------

		{
			Name: "principal_arn_nil_principal",
			Input: input{
				ac:  AuthContext{Principal: nil},
				key: "aws:PrincipalArn",
			},
			Want: EMPTY,
		},
		{
			Name: "principal_account_nil_principal",
			Input: input{
				ac:  AuthContext{Principal: nil},
				key: "aws:PrincipalAccount",
			},
			Want: EMPTY,
		},
		{
			Name: "principal_type_nil_principal",
			Input: input{
				ac:  AuthContext{Principal: nil},
				key: "aws:PrincipalType",
			},
			Want: EMPTY,
		},
		{
			Name: "principal_org_id_nil_principal",
			Input: input{
				ac:  AuthContext{Principal: nil},
				key: "aws:PrincipalOrgID",
			},
			Want: EMPTY,
		},
		{
			Name: "principal_tag_nil_principal",
			Input: input{
				ac:  AuthContext{Principal: nil},
				key: "aws:PrincipalTag/foo",
			},
			Want: EMPTY,
		},
		{
			Name: "resource_account_nil_resource",
			Input: input{
				ac:  AuthContext{Resource: nil},
				key: "aws:ResourceAccount",
			},
			Want: EMPTY,
		},
		{
			Name: "resource_org_id_nil_resource",
			Input: input{
				ac:  AuthContext{Resource: nil},
				key: "aws:ResourceOrgID",
			},
			Want: EMPTY,
		},
		{
			Name: "resource_tag_nil_resource",
			Input: input{
				ac: AuthContext{
					Action:   sar.MustLookupString("dynamodb:query"),
					Resource: nil,
				},
				key: "aws:ResourceTag/foo",
			},
			Want: EMPTY,
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
					Principal: &entities.FrozenPrincipal{
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
					Principal: &entities.FrozenPrincipal{
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
			Want: EMPTY,
		},
		{
			Name: "invalid_key",
			Input: input{
				str: "${aws:SomeStringKey}",
				ac:  AuthContext{},
			},
			Want: EMPTY,
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
				Principal: &entities.FrozenPrincipal{},
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
				Resource: &entities.FrozenResource{},
			},
			ShouldErr: true,
		},
		{
			Name: "missing_action",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{},
			},
			ShouldErr: true,
		},
		{
			Name: "resource_unexpectedly_provided",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{},
				Action:    sar.MustLookupString("sqs:listqueues"),
				Resource:  &entities.FrozenResource{},
			},
			ShouldErr: true,
		},
		{
			Name: "wrong_resource_provided",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{},
				Action:    sar.MustLookupString("sqs:getqueueurl"),
				Resource: &entities.FrozenResource{
					Arn: "arn:aws:s3:::nsiow-test",
				},
			},
			ShouldErr: true,
		},
		{
			Name: "resource_unexpectedly_missing",
			Input: AuthContext{
				Principal: &entities.FrozenPrincipal{},
				Action:    sar.MustLookupString("sqs:getqueueurl"),
			},
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i AuthContext) (any, error) {
		opts := NewOptions()
		return nil, i.Validate(opts)
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

	tests := []testlib.TestCase[input, string]{
		{
			Name: "valid_action_condition",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("s3:getobject"),
					Properties: NewBagFromMap(map[string]string{
						"s3:AuthType": "REST-HEADER",
					}),
				},
				key: "s3:authtype",
			},
			Want: "REST-HEADER",
		},
		{
			Name: "invalid_action_condition",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Properties: NewBagFromMap(map[string]string{
						"s3:AuthType": "REST-HEADER",
					}),
				},
				key: "aws:authtype",
			},
			Want: EMPTY,
		},
		{
			Name: "valid_resource_condition",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("dynamodb:query"),
					Resource: &entities.FrozenResource{
						Arn: "arn:aws:dynamodb:us-west-2:55555:table/MyTable",
						Tags: []entities.Tag{
							{
								Key:   "Foo",
								Value: "Bar",
							},
						},
					},
					Properties: NewBagFromMap(map[string]string{
						"aws:ResourceTag/Foo": "Bar",
					}),
				},
				key: "aws:ResourceTag/Foo",
			},
			Want: "Bar",
		},
		{
			Name: "another_valid_resource_condition",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("s3:getobject"),
					Resource: &entities.FrozenResource{
						Arn: "arn:aws:s3:::MyBucket/foo.txt",
						Tags: []entities.Tag{
							{
								Key:   "Foo",
								Value: "Bar",
							},
						},
					},
					Properties: NewBagFromMap(map[string]string{
						"aws:ResourceTag/Foo": "Bar",
					}),
				},
				key: "aws:ResourceTag/Foo",
			},
			Want: EMPTY,
		},
		{
			Name: "yet_another_invalid_resource_condition",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("dynamodb:query"),
					Resource: &entities.FrozenResource{
						Arn: "arn:aws:dynamodb:us-west-2:55555:table/MyTable",
						Tags: []entities.Tag{
							{
								Key:   "Foo",
								Value: "Bar",
							},
						},
					},
					Properties: NewBagFromMap(map[string]string{
						"aws:RequestTag/Foo": "Bar",
					}),
				},
				key: "aws:RequestTag/Foo",
			},
			Want: "", // aws:RequestTag not globally supported
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (string, error) {
		opts := NewOptions() // don't skip SAR validation for this
		got := i.ac.ConditionKey(i.key, opts)
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
		{
			Name: "principal_org_paths",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:listqueues"),
					Principal: &entities.FrozenPrincipal{
						Account: entities.FrozenAccount{
							OrgPaths: []string{
								"o-123/",
								"o-123/ou-1/",
								"o-123/ou-1/ou-2/",
							},
						},
					},
				},
				key: "aws:PrincipalOrgPaths",
			},
			Want: []string{
				"o-123/",
				"o-123/ou-1/",
				"o-123/ou-1/ou-2/",
			},
		},
		{
			Name: "resource_org_paths",
			Input: input{
				ac: AuthContext{
					Action: sar.MustLookupString("sqs:receivemessage"),
					Resource: &entities.FrozenResource{
						Arn: "arn:aws:sqs:us-east-1:55555:myqueue",
						Account: entities.FrozenAccount{
							OrgPaths: []string{
								"o-456/",
								"o-456/ou-a/",
							},
						},
					},
				},
				key: "aws:ResourceOrgPaths",
			},
			Want: []string{
				"o-456/",
				"o-456/ou-a/",
			},
		},
		{
			Name: "principal_org_paths_nil_principal",
			Input: input{
				ac: AuthContext{
					Principal: nil,
				},
				key: "aws:PrincipalOrgPaths",
			},
			Want: nil,
		},
		{
			Name: "resource_org_paths_nil_resource",
			Input: input{
				ac: AuthContext{
					Resource: nil,
				},
				key: "aws:ResourceOrgPaths",
			},
			Want: nil,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) ([]string, error) {
		got := i.ac.MultiKey(i.key, NewOptions())
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
			Want: EMPTY,
		},
		{
			Name: "missing_tag",
			Input: input{
				ac:  AuthContext{},
				key: "aws:PrincipalTag/color",
				tags: []entities.Tag{
					{
						Key:   "temperature",
						Value: "5",
					},
				},
			},
			Want: EMPTY,
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
					Resource: &entities.FrozenResource{
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
