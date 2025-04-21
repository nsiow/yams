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
// AWS::IAM::User
// -------------------------------------------------------------------------------------------------

type configIamUser struct {
	ConfigItem
	Configuration struct {
		AttachedManagedPolicies []policyRef    `json:"attachedManagedPolicies"`
		GroupList               []string       `json:"groupList"`
		PermissionsBoundary     boundaryRef    `json:"permissionsBoundary"`
		UserPolicies            []inlinePolicy `json:"userPolicyList"`
	} `json:"configuration"`
}

func (c *configIamUser) groupToArn(groupName string) string {
	return fmt.Sprintf("arn:aws:iam::%s:group/%s", c.AccountId, groupName)
}

func (c *configIamUser) asPrincipal() entities.Principal {
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
				return entities.Arn(x.Arn)
			}),
		Groups: common.Map(c.Configuration.GroupList,
			func(x string) entities.Arn {
				return entities.Arn(c.groupToArn(x))
			}),
		PermissionsBoundary: entities.Arn(c.Configuration.PermissionsBoundary.Arn),
	}
}

func (c *configIamUser) asResource() entities.Resource {
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

type configIamRole struct {
	ConfigItem
	Configuration struct {
		AssumeRolePolicyDocument encodedPolicy  `json:"assumeRolePolicyDocument"`
		AttachedManagedPolicies  []policyRef    `json:"attachedManagedPolicies"`
		PermissionsBoundary      boundaryRef    `json:"permissionsBoundary"`
		RolePolicies             []inlinePolicy `json:"rolePolicyList"`
	} `json:"configuration"`
}

func (c *configIamRole) asPrincipal() entities.Principal {
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
				return entities.Arn(x.Arn)
			}),
		PermissionsBoundary: entities.Arn(c.Configuration.PermissionsBoundary.Arn),
	}
}

func (c *configIamRole) asResource() entities.Resource {
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

type configIamManagedPolicy struct {
	ConfigItem
	Configuration struct {
		PolicyVersionList []struct {
			VersionId        string        `json:"versionId"`
			IsDefaultVersion bool          `json:"isDefaultVersion"`
			Document         encodedPolicy `json:"document"`
		} `json:"policyVersionList"`
	} `json:"configuration"`
}

func (c *configIamManagedPolicy) asPolicy() (entities.ManagedPolicy, error) {
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

func (c *configIamManagedPolicy) asResource() entities.Resource {
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

type configGroup struct {
	ConfigItem
	Configuration struct {
		AttachedManagedPolicies []policyRef    `json:"attachedManagedPolicies"`
		GroupPolicies           []inlinePolicy `json:"groupPolicyList"`
	} `json:"configuration"`
}

func (c *configGroup) asGroup() entities.Group {
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
				return entities.Arn(x.Arn)
			}),
	}
}

func (c *configGroup) asResource() entities.Resource {
	return entities.Resource{
		Type:      c.Type,
		AccountId: c.AccountId,
		Region:    c.Region,
		Arn:       c.Arn,
		Tags:      c.Tags,
	}
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
// AWS::S3::Bucket
// -------------------------------------------------------------------------------------------------

type configS3Bucket struct {
	ConfigItem
	SupplementaryConfiguration struct {
		BucketPolicy struct {
			PolicyText encodedPolicy `json:"policyText"`
		}
	} `json:"supplementaryConfiguration"`
}

func (c *configS3Bucket) asResource() entities.Resource {
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

type configDynamodbTable struct {
	ConfigItem
}

func (c *configDynamodbTable) asResource() entities.Resource {
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

type configKmsKey struct {
	ConfigItem
}

func (c *configKmsKey) asResource() entities.Resource {
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

type configSnsTopic struct {
	ConfigItem
	Configuration struct {
		Policy encodedPolicy
	} `json:"configuration"`
}

func (c *configSnsTopic) asResource() entities.Resource {
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

type configSqsQueue struct {
	ConfigItem
	Configuration struct {
		Policy encodedPolicy
	} `json:"configuration"`
}

func (c *configSqsQueue) asResource() entities.Resource {
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

type configAccount struct {
	ConfigItem
	Configuration struct {
		OrgId    string           `json:"orgId"`
		OrgPaths []string         `json:"orgPaths"`
		SCPs     [][]entities.Arn `json:"serviceControlPolicies"`
		RCPs     [][]entities.Arn `json:"resourceControlPolicies"`
	}
}

func (c *configAccount) asAccount() entities.Account {
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

type configSCP struct {
	ConfigItem
	Configuration struct {
		Document encodedPolicy `json:"document"`
	}
}

func (c *configSCP) asPolicy() entities.ManagedPolicy {
	return entities.ManagedPolicy{
		Type:      c.Type,
		AccountId: c.AccountId,
		Arn:       c.Arn,
		Policy:    policy.Policy(c.Configuration.Document),
	}
}

func (c *configSCP) asResource() entities.Resource {
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

type configRCP struct {
	ConfigItem
	Configuration struct {
		Document encodedPolicy `json:"document"`
	}
}

func (c *configRCP) asPolicy() entities.ManagedPolicy {
	return entities.ManagedPolicy{
		Type:      c.Type,
		AccountId: c.AccountId,
		Arn:       c.Arn,
		Policy:    policy.Policy(c.Configuration.Document),
	}
}

func (c *configRCP) asResource() entities.Resource {
	return entities.Resource{
		Type:      c.Type,
		AccountId: c.AccountId,
		Region:    c.Region,
		Arn:       c.Arn,
		Tags:      c.Tags,
	}
}
