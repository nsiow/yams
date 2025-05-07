package awsconfig

import (
	"fmt"

	"github.com/nsiow/yams/internal/common"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// -------------------------------------------------------------------------------------------------
// Shared fragments
// -------------------------------------------------------------------------------------------------

type policyRef struct {
	Arn  string `json:"policyArn"`
	Name string `json:"policyName"`
}

type boundaryRef struct {
	Arn  string `json:"permissionsBoundaryArn"`
	Name string `json:"permissionsBoundaryName"`
}

type inlinePolicy struct {
	Name     string        `json:"policyName"`
	Document encodedPolicy `json:"policyDocument"`
}

// -------------------------------------------------------------------------------------------------
// Generic resource
// -------------------------------------------------------------------------------------------------

type genericResource struct {
	ConfigItem
}

func (c *genericResource) asResource() entities.Resource {
	return entities.Resource{
		Type:      c.Type,
		AccountId: c.AccountId,
		Region:    c.Region,
		Arn:       c.Arn,
		Tags:      c.Tags,
	}
}

// -------------------------------------------------------------------------------------------------
// AWS::IAM::User
// -------------------------------------------------------------------------------------------------

type IamUser struct {
	ConfigItem
	Configuration struct {
		AttachedManagedPolicies []policyRef    `json:"attachedManagedPolicies"`
		GroupList               []string       `json:"groupList"`
		PermissionsBoundary     boundaryRef    `json:"permissionsBoundary"`
		UserPolicies            []inlinePolicy `json:"userPolicyList"`
	} `json:"configuration"`
}

func (c *IamUser) groupToArn(groupName string) string {
	return fmt.Sprintf("arn:aws:iam::%s:group/%s", c.AccountId, groupName)
}

func (c *IamUser) asPrincipal() entities.Principal {
	return entities.Principal{
		Type:      c.Type,
		AccountId: c.AccountId,
		Arn:       c.Arn,
		Tags:      c.Tags,
		InlinePolicies: common.Map(c.Configuration.UserPolicies,
			func(x inlinePolicy) policy.Policy {
				return policy.Policy(x.Document)
			}),
		AttachedPolicies: common.Map(c.Configuration.AttachedManagedPolicies,
			func(x policyRef) entities.Arn {
				return x.Arn
			}),
		Groups: common.Map(c.Configuration.GroupList,
			func(x string) entities.Arn {
				return c.groupToArn(x)
			}),
		PermissionsBoundary: c.Configuration.PermissionsBoundary.Arn,
	}
}

func (c *IamUser) asResource() entities.Resource {
	return entities.Resource{
		Type:      c.Type,
		AccountId: c.AccountId,
		Region:    c.Region,
		Arn:       c.Arn,
		Tags:      c.Tags,
	}
}

// -------------------------------------------------------------------------------------------------
// AWS::IAM::Role
// -------------------------------------------------------------------------------------------------

type IamRole struct {
	ConfigItem
	Configuration struct {
		AssumeRolePolicyDocument encodedPolicy  `json:"assumeRolePolicyDocument"`
		AttachedManagedPolicies  []policyRef    `json:"attachedManagedPolicies"`
		PermissionsBoundary      boundaryRef    `json:"permissionsBoundary"`
		RolePolicies             []inlinePolicy `json:"rolePolicyList"`
	} `json:"configuration"`
}

func (c *IamRole) asPrincipal() entities.Principal {
	return entities.Principal{
		Type:      c.Type,
		AccountId: c.AccountId,
		Arn:       c.Arn,
		Tags:      c.Tags,
		InlinePolicies: common.Map(c.Configuration.RolePolicies,
			func(x inlinePolicy) policy.Policy {
				return policy.Policy(x.Document)
			}),
		AttachedPolicies: common.Map(c.Configuration.AttachedManagedPolicies,
			func(x policyRef) entities.Arn {
				return x.Arn
			}),
		PermissionsBoundary: c.Configuration.PermissionsBoundary.Arn,
	}
}

func (c *IamRole) asResource() entities.Resource {
	return entities.Resource{
		Type:      c.Type,
		AccountId: c.AccountId,
		Region:    c.Region,
		Arn:       c.Arn,
		Tags:      c.Tags,
		Policy:    policy.Policy(c.Configuration.AssumeRolePolicyDocument),
	}
}

// -------------------------------------------------------------------------------------------------
// AWS::IAM::Policy
// -------------------------------------------------------------------------------------------------

type IamPolicy struct {
	ConfigItem
	Configuration struct {
		PolicyVersionList []struct {
			VersionId        string        `json:"versionId"`
			IsDefaultVersion bool          `json:"isDefaultVersion"`
			Document         encodedPolicy `json:"document"`
		} `json:"policyVersionList"`
	} `json:"configuration"`
}

func (c *IamPolicy) asPolicy() (entities.ManagedPolicy, error) {
	for _, pv := range c.Configuration.PolicyVersionList {
		if pv.IsDefaultVersion {
			return entities.ManagedPolicy{
				Type:      c.Type,
				AccountId: c.AccountId,
				Arn:       c.Arn,
				Policy:    policy.Policy(pv.Document),
			}, nil
		}
	}

	return entities.ManagedPolicy{}, fmt.Errorf("unable to find default policy version for: %s", c.Arn)
}

func (c *IamPolicy) asResource() entities.Resource {
	return entities.Resource{
		Type:      c.Type,
		AccountId: c.AccountId,
		Region:    c.Region,
		Arn:       c.Arn,
		Tags:      c.Tags,
	}
}

// -------------------------------------------------------------------------------------------------
// AWS::IAM::Group
// -------------------------------------------------------------------------------------------------

type IamGroup struct {
	ConfigItem
	Configuration struct {
		AttachedManagedPolicies []policyRef    `json:"attachedManagedPolicies"`
		GroupPolicies           []inlinePolicy `json:"groupPolicyList"`
	} `json:"configuration"`
}

func (c *IamGroup) asGroup() entities.Group {
	return entities.Group{
		Type:      c.Type,
		AccountId: c.AccountId,
		Arn:       c.Arn,
		InlinePolicies: common.Map(c.Configuration.GroupPolicies,
			func(x inlinePolicy) policy.Policy {
				return policy.Policy(x.Document)
			}),
		AttachedPolicies: common.Map(c.Configuration.AttachedManagedPolicies,
			func(x policyRef) entities.Arn {
				return x.Arn
			}),
	}
}

func (c *IamGroup) asResource() entities.Resource {
	return entities.Resource{
		Type:      c.Type,
		AccountId: c.AccountId,
		Region:    c.Region,
		Arn:       c.Arn,
		Tags:      c.Tags,
	}
}

// -------------------------------------------------------------------------------------------------
// AWS::S3::Bucket
// -------------------------------------------------------------------------------------------------

type S3Bucket struct {
	ConfigItem
	SupplementaryConfiguration struct {
		BucketPolicy struct {
			PolicyText encodedPolicy `json:"policyText"`
		}
	} `json:"supplementaryConfiguration"`
}

func (c *S3Bucket) asResource() entities.Resource {
	return entities.Resource{
		Type:      c.Type,
		AccountId: c.AccountId,
		Region:    c.Region,
		Arn:       c.Arn,
		Tags:      c.Tags,
		Policy:    policy.Policy(c.SupplementaryConfiguration.BucketPolicy.PolicyText),
	}
}

// -------------------------------------------------------------------------------------------------
// AWS::DynamoDB::Table
// -------------------------------------------------------------------------------------------------

type DynamodbTable struct {
	ConfigItem
}

func (c *DynamodbTable) asResource() entities.Resource {
	return entities.Resource{
		Type:      c.Type,
		AccountId: c.AccountId,
		Region:    c.Region,
		Arn:       c.Arn,
		Tags:      c.Tags,
		Policy:    policy.Policy{}, // TODO(nsiow) implement DDB table policy support
	}
}

// -------------------------------------------------------------------------------------------------
// AWS::KMS::Key
// -------------------------------------------------------------------------------------------------

type KmsKey struct {
	ConfigItem
}

func (c *KmsKey) asResource() entities.Resource {
	return entities.Resource{
		Type:      c.Type,
		AccountId: c.AccountId,
		Region:    c.Region,
		Arn:       c.Arn,
		Tags:      c.Tags,
		Policy:    policy.Policy{}, // TODO(nsiow) implement KMS key policy support
	}
}

// -------------------------------------------------------------------------------------------------
// AWS::SNS::Topic
// -------------------------------------------------------------------------------------------------

type SnsTopic struct {
	ConfigItem
	Configuration struct {
		Policy encodedPolicy
	} `json:"configuration"`
}

func (c *SnsTopic) asResource() entities.Resource {
	return entities.Resource{
		Type:      c.Type,
		AccountId: c.AccountId,
		Region:    c.Region,
		Arn:       c.Arn,
		Tags:      c.Tags,
		Policy:    policy.Policy(c.Configuration.Policy),
	}
}

// -------------------------------------------------------------------------------------------------
// AWS::SQS::Queue
// -------------------------------------------------------------------------------------------------

type SqsQueue struct {
	ConfigItem
	Configuration struct {
		Policy encodedPolicy
	} `json:"configuration"`
}

func (c *SqsQueue) asResource() entities.Resource {
	return entities.Resource{
		Type:      c.Type,
		AccountId: c.AccountId,
		Region:    c.Region,
		Arn:       c.Arn,
		Tags:      c.Tags,
		Policy:    policy.Policy(c.Configuration.Policy),
	}
}

// -------------------------------------------------------------------------------------------------
// Yams::Account
// -------------------------------------------------------------------------------------------------

type Account struct {
	ConfigItem
	Configuration AccountConfiguration `json:"configuration"`
}

type AccountConfiguration struct {
	OrgId    string           `json:"orgId"`
	OrgPaths []string         `json:"orgPaths"`
	SCPs     [][]entities.Arn `json:"serviceControlPolicies"`
	RCPs     [][]entities.Arn `json:"resourceControlPolicies"`
}

func (c *Account) asAccount() entities.Account {
	return entities.Account{
		Id:       c.AccountId,
		OrgId:    c.Configuration.OrgId,
		OrgPaths: c.Configuration.OrgPaths,
		SCPs:     c.Configuration.SCPs,
		RCPs:     c.Configuration.RCPs,
	}
}

// -------------------------------------------------------------------------------------------------
// Yams::ServiceControlPolicy
// -------------------------------------------------------------------------------------------------

type SCP struct {
	ConfigItem
	Configuration SCPConfiguration `json:"configuration"`
}

type SCPConfiguration struct {
	Document encodedPolicy `json:"document"`
}

func (c *SCP) asPolicy() entities.ManagedPolicy {
	return entities.ManagedPolicy{
		Type:      c.Type,
		AccountId: c.AccountId,
		Arn:       c.Arn,
		Policy:    policy.Policy(c.Configuration.Document),
	}
}

func (c *SCP) asResource() entities.Resource {
	return entities.Resource{
		Type:      c.Type,
		AccountId: c.AccountId,
		Region:    c.Region,
		Arn:       c.Arn,
		Tags:      c.Tags,
	}
}

// -------------------------------------------------------------------------------------------------
// Yams::ResourceControlPolicy
// -------------------------------------------------------------------------------------------------

type RCP struct {
	ConfigItem
	Configuration RCPConfiguration `json:"configuration"`
}

type RCPConfiguration struct {
	Document encodedPolicy `json:"document"`
}

func (c *RCP) asPolicy() entities.ManagedPolicy {
	return entities.ManagedPolicy{
		Type:      c.Type,
		AccountId: c.AccountId,
		Arn:       c.Arn,
		Policy:    policy.Policy(c.Configuration.Document),
	}
}

func (c *RCP) asResource() entities.Resource {
	return entities.Resource{
		Type:      c.Type,
		AccountId: c.AccountId,
		Region:    c.Region,
		Arn:       c.Arn,
		Tags:      c.Tags,
	}
}
