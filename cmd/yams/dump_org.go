package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/nsiow/yams/pkg/loaders/awsconfig"
	"github.com/nsiow/yams/pkg/policy"
)

// -------------------------------------------------------------------------------------------------
// Dumping logic
// -------------------------------------------------------------------------------------------------

func dumpOrg() (string, error) {
	ctx := context.Background()
	client, err := orgClient(ctx)
	if err != nil {
		return "", fmt.Errorf("error creating org client: %w", err)
	}

	accounts, err := walk(ctx, client)
	if err != nil {
		return "", fmt.Errorf("error walking accounts: %w", err)
	}

	scps, err := describeScps(ctx, client)
	if err != nil {
		return "", fmt.Errorf("error walking SCPs: %w", err)
	}

	rcps, err := describeRcps(ctx, client)
	if err != nil {
		return "", fmt.Errorf("error walking RCPs: %w", err)
	}

	var orgEntities []any
	for _, entity := range accounts {
		orgEntities = append(orgEntities, entity)
	}
	for _, entity := range scps {
		orgEntities = append(orgEntities, entity)
	}
	for _, entity := range rcps {
		orgEntities = append(orgEntities, entity)
	}

	asJson, err := json.Marshal(orgEntities)
	if err != nil {
		return "", fmt.Errorf("error marshalling dump as json: %w", err)
	}

	return string(asJson), nil
}

// -------------------------------------------------------------------------------------------------
// Organizations API stuff
// -------------------------------------------------------------------------------------------------

func isAccount(id string) bool {
	_, err := strconv.Atoi(id)
	return err == nil
}

// orgClient creates and returns a new AWS Organizations SDK client using the provided options
func orgClient(ctx context.Context) (*organizations.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := organizations.NewFromConfig(cfg)
	return client, nil
}

func orgId(ctx context.Context, client *organizations.Client) (string, error) {
	resp, err := client.DescribeOrganization(ctx, &organizations.DescribeOrganizationInput{})
	if err != nil {
		return "", err
	}

	slog.Debug("found org id", "id", *resp.Organization.Id)
	return *resp.Organization.Id, nil
}

func orgRoot(ctx context.Context, client *organizations.Client) (string, error) {
	resp, err := client.ListRoots(ctx, &organizations.ListRootsInput{})
	if err != nil {
		return "", err
	}

	if len(resp.Roots) != 1 {
		return "", fmt.Errorf("unexpected number of roots: %d (%v)", len(resp.Roots), resp.Roots)
	}

	slog.Debug("found org root", "id", *resp.Roots[0].Id)
	return *resp.Roots[0].Id, nil
}

func orgPaths(orgId string, path []string) []string {
	var paths []string

	segment := orgId + "/"

	for _, p := range path {
		if !isAccount(p) {
			segment += p + "/"
			paths = append(paths, segment)
		}
	}

	slog.Debug("calculated orgpaths", "input", path, "paths", paths)
	return paths
}

func walk(ctx context.Context, client *organizations.Client) ([]awsconfig.Account, error) {
	id, err := orgId(ctx, client)
	if err != nil {
		return nil, err
	}

	root, err := orgRoot(ctx, client)
	if err != nil {
		return nil, err
	}

	var accounts []awsconfig.Account
	slog.Debug("preparing to walk org tree")
	_walk(ctx, client, id, &accounts, []string{root})
	slog.Debug("finished walking org tree")

	return accounts, nil
}

func _walk(
	ctx context.Context,
	client *organizations.Client,
	orgId string,
	accounts *[]awsconfig.Account, path []string) error {
	node := path[len(path)-1]

	if isAccount(node) {
		slog.Debug("found account", "id", node)
		a, err := makeAccount(ctx, client, orgId, path, node)
		if err != nil {
			return err
		}

		*accounts = append(*accounts, *a)
		return nil
	}

	var nextToken *string
	childTypes := []types.ChildType{types.ChildTypeAccount, types.ChildTypeOrganizationalUnit}

	for _, childType := range childTypes {
		for {
			resp, err := client.ListChildren(ctx, &organizations.ListChildrenInput{
				ChildType: childType,
				ParentId:  &node,
				NextToken: nextToken,
			})
			if err != nil {
				return err
			}
			slog.Debug("listed children", "parent", node, "children", resp.Children)

			for _, child := range resp.Children {
				_walk(ctx, client, orgId, accounts, append(path, *child.Id))
			}

			if resp.NextToken == nil {
				break
			}
			nextToken = resp.NextToken
		}
	}

	return nil
}

func describePolicies(
	ctx context.Context,
	client *organizations.Client,
	policyType types.PolicyType) ([]*organizations.DescribePolicyOutput, error) {

	var policies []*organizations.DescribePolicyOutput
	var nextToken *string

	for {
		resp, err := client.ListPolicies(ctx, &organizations.ListPoliciesInput{
			Filter:    policyType,
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}
		slog.Debug("found policies for type", "type", policyType, "numPolicies", len(resp.Policies))

		for _, policySummary := range resp.Policies {
			resp, err := client.DescribePolicy(ctx, &organizations.DescribePolicyInput{
				PolicyId: policySummary.Id,
			})
			if err != nil {
				return nil, err
			}

			policies = append(policies, resp)
		}

		if resp.NextToken == nil {
			break
		}
		nextToken = resp.NextToken
	}

	return policies, nil
}

func describeScps(ctx context.Context, client *organizations.Client) ([]awsconfig.SCP, error) {
	raw, err := describePolicies(ctx, client, types.PolicyTypeServiceControlPolicy)
	if err != nil {
		return nil, err
	}

	var structured []awsconfig.SCP
	for _, rawPolicy := range raw {
		policyDocument, err := policy.FromJsonString(*rawPolicy.Policy.Content)
		if err != nil {
			return nil, err
		}

		s := awsconfig.SCP{
			ConfigItem: awsconfig.ConfigItem{
				Type: "Yams::ServiceControlPolicy",
				Arn:  *rawPolicy.Policy.PolicySummary.Arn,
			},
			Configuration: awsconfig.SCPConfiguration{
				Document: awsconfig.EncodedPolicy(policyDocument),
			},
		}
		structured = append(structured, s)
	}

	return structured, nil
}

func describeRcps(ctx context.Context, client *organizations.Client) ([]awsconfig.SCP, error) {
	raw, err := describePolicies(ctx, client, types.PolicyTypeResourceControlPolicy)
	if err != nil {
		return nil, err
	}

	var structured []awsconfig.SCP
	for _, rawPolicy := range raw {
		policyDocument, err := policy.FromJsonString(*rawPolicy.Policy.Content)
		if err != nil {
			return nil, err
		}

		s := awsconfig.SCP{
			ConfigItem: awsconfig.ConfigItem{
				Type: "Yams::ResourceControlPolicy",
				Arn:  *rawPolicy.Policy.PolicySummary.Arn,
			},
			Configuration: awsconfig.SCPConfiguration{
				Document: awsconfig.EncodedPolicy(policyDocument),
			},
		}
		structured = append(structured, s)
	}

	return structured, nil
}

func listPoliciesForNodes(
	ctx context.Context,
	client *organizations.Client,
	policyType types.PolicyType,
	nodeId string) ([]string, error) {

	slog.Debug("fetching policies for node", "id", nodeId)

	var policies []string
	var nextToken *string

	for {
		resp, err := client.ListPoliciesForTarget(ctx, &organizations.ListPoliciesForTargetInput{
			TargetId:  &nodeId,
			Filter:    policyType,
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}
		slog.Debug("found policies for node", "id", nodeId, "numPolicies", len(resp.Policies))

		for _, policySummary := range resp.Policies {
			policies = append(policies, *policySummary.Arn)
		}

		if resp.NextToken == nil {
			break
		}
		nextToken = resp.NextToken
	}

	return policies, nil
}

func listScpsForNode(
	ctx context.Context,
	client *organizations.Client,
	nodeId string) ([]string, error) {

	return listPoliciesForNodes(ctx, client, types.PolicyTypeServiceControlPolicy, nodeId)

}

func listRcpsForNode(
	ctx context.Context,
	client *organizations.Client,
	nodeId string) ([]string, error) {

	return listPoliciesForNodes(ctx, client, types.PolicyTypeResourceControlPolicy, nodeId)

}

func describeAccount(
	ctx context.Context,
	client *organizations.Client,
	nodeId string) (*organizations.DescribeAccountOutput, error) {

	return client.DescribeAccount(ctx, &organizations.DescribeAccountInput{
		AccountId: &nodeId,
	})

}

func describeOu(
	ctx context.Context,
	client *organizations.Client,
	nodeId string) (*organizations.DescribeOrganizationalUnitOutput, error) {

	return client.DescribeOrganizationalUnit(ctx, &organizations.DescribeOrganizationalUnitInput{
		OrganizationalUnitId: &nodeId,
	})

}

func orgNode(
	ctx context.Context,
	client *organizations.Client,
	path []string,
	nodeId string) (*awsconfig.OrgNode, error) {

	var id, arn, name, nodeType string
	if isAccount(nodeId) {
		nodeType = "ACCOUNT"

		summary, err := describeAccount(ctx, client, nodeId)
		if err != nil {
			return nil, err
		}

		id = *summary.Account.Id
		arn = *summary.Account.Arn
		name = *summary.Account.Name
	} else {
		nodeType = "ORGANIZATIONAL_UNIT"

		summary, err := describeOu(ctx, client, nodeId)
		if err != nil {
			return nil, err
		}

		id = *summary.OrganizationalUnit.Id
		arn = *summary.OrganizationalUnit.Arn
		name = *summary.OrganizationalUnit.Name
	}

	scps, err := listScpsForNode(ctx, client, nodeId)
	if err != nil {
		return nil, err
	}

	rcps, err := listRcpsForNode(ctx, client, nodeId)
	if err != nil {
		return nil, err
	}

	return &awsconfig.OrgNode{
		Id:   id,
		Type: nodeType,
		Arn:  arn,
		Name: name,
		SCPs: scps,
		RCPs: rcps,
	}, nil
}

func orgNodes(
	ctx context.Context,
	client *organizations.Client,
	path []string) ([]awsconfig.OrgNode, error) {

	nodes := make([]awsconfig.OrgNode, len(path))
	for i, nodeId := range path {
		node, err := orgNode(ctx, client, path, nodeId)
		if err != nil {
			return nil, err
		}

		nodes[i] = *node
	}

	return nodes, nil
}

func makeAccount(
	ctx context.Context,
	client *organizations.Client,
	orgId string,
	path []string,
	node string) (*awsconfig.Account, error) {

	nodes, err := orgNodes(ctx, client, path)
	if err != nil {
		return nil, err
	}

	return &awsconfig.Account{
		ConfigItem: awsconfig.ConfigItem{
			Type:      "Yams::Account",
			AccountId: node,
		},
		Configuration: awsconfig.AccountConfiguration{
			OrgId:    orgId,
			OrgPaths: orgPaths(orgId, path),
			OrgNodes: nodes,
		},
	}, nil
}
