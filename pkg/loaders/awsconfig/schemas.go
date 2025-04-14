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

type configUser struct {
	ConfigItem
	AttachedManagedPolicies []policyRef    `json:"attachedManagedPolicies"`
	GroupList               []string       `json:"groupList"`
	PermissionsBoundary     boundaryRef    `json:"permissionsBoundary"`
	UserPolicies            []inlinePolicy `json:"userPolicyList"`
}

func (c *configUser) toEntity() (entities.Principal, error) {
	return entities.Principal{
		Type:      c.Type,
		AccountId: c.AccountId,
		Arn:       c.Arn,
		Tags:      c.Tags,
		InlinePolicies: common.Map(c.UserPolicies, func(x inlinePolicy) policy.Policy {
			return policy.Policy(x.Document)
		}),
		AttachedPolicies: common.Map(c.AttachedManagedPolicies, func(x policyRef) entities.Arn {
			return entities.Arn(x.Arn)
		}),
		// FIXME(nsiow) make a helper function for this
		Groups: common.Map(c.GroupList, func(x string) entities.Arn {
			groupArn := fmt.Sprintf("arn:aws:iam::%s:group/%s", c.AccountId, x)
			return entities.Arn(groupArn)
		}),
		PermissionsBoundary: entities.Arn(c.PermissionsBoundary.Arn),
	}, nil
}

// -------------------------------------------------------------------------------------------------
// AWS::IAM::Policy
// -------------------------------------------------------------------------------------------------

type configManagedPolicy struct {
	ConfigItem
	PolicyVersionList []struct {
		VersionId        string        `json:"versionId"`
		IsDefaultVersion bool          `json:"isDefaultVersion"`
		Document         encodedPolicy `json:"document"`
	} `json:"policyVersionList"`
}

func (c *configManagedPolicy) toEntity() (entities.Policy, error) {
	for _, pv := range c.PolicyVersionList {
		if pv.IsDefaultVersion {
			return entities.Policy{
				Type:      c.Type,
				AccountId: c.AccountId,
				Arn:       c.Arn,
				Policy:    policy.Policy(pv.Document),
			}, nil
		}
	}

	return entities.Policy{}, fmt.Errorf("unable to find default policy version for: %s", c.Arn)
}

// -------------------------------------------------------------------------------------------------
// AWS::IAM::Group
// -------------------------------------------------------------------------------------------------

type configGroup struct {
	ConfigItem
	AttachedManagedPolicies []policyRef `json:"attachedManagedPolicies"`
}

func (c *configGroup) toEntity() (entities.Group, error) {
	arns := []entities.Arn{}
	for _, policyRef := range c.AttachedManagedPolicies {
		arns = append(arns, entities.Arn(policyRef.Arn))
	}

	return entities.Group{
		Type:      c.Type,
		AccountId: c.AccountId,
		Arn:       c.Arn,
		Policies: common.Map(c.AttachedManagedPolicies, func(x policyRef) entities.Arn {
			return entities.Arn(x.Arn)
		}),
	}, nil
}

// -------------------------------------------------------------------------------------------------
// Yams::Account
// -------------------------------------------------------------------------------------------------

type configAccount struct {
	ConfigItem
	OrgId    string           `json:"orgId"`
	OrgPaths []string         `json:"orgPaths"`
	SCPs     [][]entities.Arn `json:"serviceControlPolicies"`
}

func (c *configAccount) toEntity() (entities.Account, error) {
	return entities.Account{
		Id:       c.AccountId,
		OrgId:    c.OrgId,
		OrgPaths: c.OrgPaths,
		SCPs:     c.SCPs,
	}, nil
}

// -------------------------------------------------------------------------------------------------
// Yams::ServiceControlPolicy
// -------------------------------------------------------------------------------------------------

type configSCP struct {
	ConfigItem
	Document encodedPolicy `json:"document"`
}

func (c *configSCP) toEntity() (entities.Policy, error) {
	return entities.Policy{
		Type:      c.Type,
		AccountId: c.AccountId,
		Arn:       c.Arn,
		Policy:    policy.Policy(c.Document),
	}, nil
}
