package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
)

func TestIsStrictCall(t *testing.T) {
	tests := []testlib.TestCase[AuthContext, bool]{
		{
			Name:  "empty_subject",
			Input: AuthContext{},
			Want:  false,
		},
		{
			Name: "non_strict_call",
			Input: AuthContext{
				Action: sar.MustLookupString("s3:getobject"),
			},
			Want: false,
		},
		{
			Name: "sts_assume_role",
			Input: AuthContext{
				Action:    sar.MustLookupString("sts:assumerole"),
				Principal: &entities.Principal{},
				Resource:  &entities.Resource{},
			},
			Want: true,
		},
		{
			Name: "kms_plus_key",
			Input: AuthContext{
				Action:    sar.MustLookupString("kms:decrypt"),
				Principal: &entities.Principal{},
				Resource: &entities.Resource{
					Arn:  "arn:aws:kms:us-west-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab",
					Type: "AWS::KMS::Key",
				},
			},
			Want: true,
		},
		{
			Name: "kms_sans_key",
			Input: AuthContext{
				Action:    sar.MustLookupString("kms:decrypt"),
				Principal: &entities.Principal{},
				Resource: &entities.Resource{
					Arn:  "arn:aws:kms:us-west-2:111122223333:alias/ExampleAlias",
					Type: "AWS::KMS::Alias",
				},
			},
			Want: false,
		},
	}

	testlib.RunTestSuite(t, tests, func(i AuthContext) (bool, error) {
		subj := newSubject(&i, TestingSimulationOptions)
		return isStrictCall(subj), nil
	})
}

// }
//
// func TestSameAccountExplicitPrincipalCase(t *testing.T) {
// 	tests := []testlib.TestCase[AuthContext, bool]{
// 		{
// 			Name:  "empty_auth_context",
// 			Input: AuthContext{},
// 			Want:  false,
// 		},
// 		{
// 			Name: "same_account_explicit_principal",
// 			Input: AuthContext{
// 				Action: sar.MustLookupString("s3:getobject"),
// 				Principal: &entities.Principal{
// 					Type: "AWS::IAM::Role",
// 					Arn:  "arn:aws:iam::55555:role/MyRole",
// 				},
// 				Resource: &entities.Resource{
// 					Arn:  "arn:aws:s3:::nsiow-test",
// 					Type: "AWS::S3::Bucket",
// 					Policy: policy.Policy{
// 						Statement: []policy.Statement{
// 							{
// 								Sid: "test_statement",
// 								Principal: policy.Principal{
// 									AWS: policy.Value{"arn:aws:iam::55555:role/MyRole"},
// 								},
// 								Effect:   "Allow",
// 								Action:   []string{"s3:getobject"},
// 								Resource: []string{"arn:aws:s3:::nsiow-test/*"},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			Want: true,
// 		},
// 		{
// 			Name: "same_account_explicit_principal_unrelated_actions",
// 			Input: AuthContext{
// 				Action: sar.MustLookupString("s3:getobject"),
// 				Principal: &entities.Principal{
// 					Type: "AWS::IAM::Role",
// 					Arn:  "arn:aws:iam::55555:role/MyRole",
// 				},
// 				Resource: &entities.Resource{
// 					Arn:  "arn:aws:s3:::nsiow-test",
// 					Type: "AWS::S3::Bucket",
// 					Policy: policy.Policy{
// 						Statement: []policy.Statement{
// 							{
// 								Sid: "test_statement",
// 								Principal: policy.Principal{
// 									AWS: policy.Value{"arn:aws:iam::55555:role/MyRole"},
// 								},
// 								Effect:   "Allow",
// 								Action:   []string{"s3:listbucket"},
// 								Resource: []string{"arn:aws:s3:::nsiow-test"},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			Want: false,
// 		},
// 		{
// 			Name: "same_account_principal_star",
// 			Input: AuthContext{
// 				Action: sar.MustLookupString("s3:getobject"),
// 				Principal: &entities.Principal{
// 					Type: "AWS::IAM::Role",
// 					Arn:  "arn:aws:iam::55555:role/MyRole",
// 				},
// 				Resource: &entities.Resource{
// 					Arn:  "arn:aws:s3:::nsiow-test",
// 					Type: "AWS::S3::Bucket",
// 					Policy: policy.Policy{
// 						Statement: []policy.Statement{
// 							{
// 								Sid: "test_statement",
// 								Principal: policy.Principal{
// 									AWS: policy.Value{"*"},
// 								},
// 								Effect:   "Allow",
// 								Action:   []string{"s3:getobject"},
// 								Resource: []string{"arn:aws:s3:::nsiow-test/*"},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			Want: false,
// 		},
// 	}
//
// 	testlib.RunTestSuite(t, tests, func(i AuthContext) (bool, error) {
// 		subj := newSubject(&i, TestingSimulationOptions)
// 		access, err := evalSameAccountExplicitPrincipalCase(subj)
// 		return access.Allowed(), err
// 	})
// }
