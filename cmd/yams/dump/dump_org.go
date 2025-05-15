package dump

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/nsiow/yams/cmd/yams/cli"
	"github.com/nsiow/yams/pkg/loaders/awsconfig"
	"github.com/nsiow/yams/pkg/policy"
)

// -------------------------------------------------------------------------------------------------
// Dumping logic
// -------------------------------------------------------------------------------------------------

func Org(_ *cli.Flags) (string, error) {
	ctx := context.Background()
	client, err := orgClient(ctx)
	if err != nil {
		return "", fmt.Errorf("error creating org client: %w", err)
	}

	cache := make(map[string]any)

	accounts, err := walk(ctx, client, cache)
	if err != nil {
		return "", fmt.Errorf("error walking accounts: %w", err)
	}

	scps, err := describeScps(ctx, client, cache)
	if err != nil {
		return "", fmt.Errorf("error walking SCPs: %w", err)
	}

	rcps, err := describeRcps(ctx, client, cache)
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
// Org walking
// -------------------------------------------------------------------------------------------------

func walk(
	ctx context.Context,
	client *organizations.Client,
	cache map[string]any) ([]awsconfig.Account, error) {
	org, err := describeOrg(ctx, client, cache)
	if err != nil {
		return nil, err
	}

	root, err := describeOrgRoot(ctx, client, cache)
	if err != nil {
		return nil, err
	}

	slog.Debug("preparing to walk org tree")
	accounts, err := _walk(ctx, client, cache, *org.Id, []string{*root.Id})
	if err != nil {
		return nil, err
	}

	slog.Debug("finished walking org tree")
	return accounts, err
}

func _walk(
	ctx context.Context,
	client *organizations.Client,
	cache map[string]any,
	orgId string,
	path []string) ([]awsconfig.Account, error) {

	node := path[len(path)-1]

	if isAccount(node) {
		slog.Debug("found account", "id", node)
		a, err := makeAccount(ctx, client, cache, orgId, path, node)
		if err != nil {
			return nil, err
		}

		return []awsconfig.Account{*a}, nil
	}

	var accounts []awsconfig.Account
	childTypes := []types.ChildType{types.ChildTypeAccount, types.ChildTypeOrganizationalUnit}

	for _, childType := range childTypes {
		var nextToken *string

		for {
			resp, err := client.ListChildren(ctx, &organizations.ListChildrenInput{
				ChildType: childType,
				ParentId:  &node,
				NextToken: nextToken,
			})
			if err != nil {
				return nil, err
			}
			slog.Debug("listed children", "parent", node, "children", resp.Children)

			for _, child := range resp.Children {
				childAccounts, err := _walk(ctx, client, cache, orgId, append(path, *child.Id))
				if err != nil {
					return nil, err
				}

				accounts = slices.Concat(accounts, childAccounts)
			}

			if resp.NextToken == nil {
				break
			}
			nextToken = resp.NextToken
		}
	}

	return accounts, nil
}

// -------------------------------------------------------------------------------------------------
// Organizations API stuff
// -------------------------------------------------------------------------------------------------

// orgClient creates and returns a new AWS Organizations SDK client using the provided options
func orgClient(ctx context.Context) (*organizations.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	client := organizations.NewFromConfig(cfg)
	return client, nil
}

func describeOrg(
	ctx context.Context,
	client *organizations.Client,
	cache map[string]any) (*types.Organization, error) {

	key := "describeOrg/"
	if cached, ok := cache[key].(*types.Organization); ok {
		return cached, nil
	}

	resp, err := client.DescribeOrganization(ctx, &organizations.DescribeOrganizationInput{})
	if err != nil {
		return nil, err
	}

	slog.Debug("found org id", "id", *resp.Organization.Id)
	cache[key] = resp.Organization
	return resp.Organization, nil
}

func describeOrgRoot(
	ctx context.Context,
	client *organizations.Client,
	cache map[string]any) (*types.Root, error) {

	key := "describeOrgRoot/"
	if cached, ok := cache[key].(*types.Root); ok {
		return cached, nil
	}

	resp, err := client.ListRoots(ctx, &organizations.ListRootsInput{})
	if err != nil {
		return nil, err
	}

	if len(resp.Roots) != 1 {
		return nil, fmt.Errorf("unexpected number of roots: %d (%v)", len(resp.Roots), resp.Roots)
	}

	slog.Debug("found org root", "id", *resp.Roots[0].Id)
	cache[key] = &resp.Roots[0]
	return &resp.Roots[0], nil
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
	cache map[string]any,
	nodeId string) ([]string, error) {

	key := "listScpsForNode/" + nodeId
	if cached, ok := cache[key].([]string); ok {
		return cached, nil
	}

	policies, err := listPoliciesForNodes(ctx, client, types.PolicyTypeServiceControlPolicy, nodeId)
	if err != nil {
		return nil, err
	}

	cache[key] = policies
	return policies, nil
}

func listRcpsForNode(
	ctx context.Context,
	client *organizations.Client,
	cache map[string]any,
	nodeId string) ([]string, error) {

	key := "listRcpsForNode/" + nodeId
	if cached, ok := cache[key].([]string); ok {
		return cached, nil
	}

	policies, err := listPoliciesForNodes(ctx, client, types.PolicyTypeResourceControlPolicy, nodeId)
	if err != nil {
		return nil, err
	}

	cache[key] = policies
	return policies, nil
}

func describeAccount(
	ctx context.Context,
	client *organizations.Client,
	cache map[string]any,
	nodeId string) (*organizations.DescribeAccountOutput, error) {

	key := "describeAccount/" + nodeId
	if cached, ok := cache[key].(*organizations.DescribeAccountOutput); ok {
		return cached, nil
	}

	slog.Debug("calling organizations describe-account", "id", nodeId)
	resp, err := client.DescribeAccount(ctx, &organizations.DescribeAccountInput{
		AccountId: &nodeId,
	})
	if err != nil {
		return nil, err
	}

	cache[key] = resp
	return resp, nil
}

func describeOu(
	ctx context.Context,
	client *organizations.Client,
	cache map[string]any,
	nodeId string) (*organizations.DescribeOrganizationalUnitOutput, error) {

	key := "describeOu/" + nodeId
	if cached, ok := cache[key].(*organizations.DescribeOrganizationalUnitOutput); ok {
		return cached, nil
	}

	slog.Debug("calling organizations describe-organizational-unit", "id", nodeId)
	resp, err := client.DescribeOrganizationalUnit(ctx, &organizations.DescribeOrganizationalUnitInput{
		OrganizationalUnitId: &nodeId,
	})
	if err != nil {
		return nil, err
	}

	cache[key] = resp
	return resp, nil
}

// -------------------------------------------------------------------------------------------------
// Helper functions
// -------------------------------------------------------------------------------------------------

func isAccount(id string) bool {
	_, err := strconv.Atoi(id)
	return err == nil
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

func orgNode(
	ctx context.Context,
	client *organizations.Client,
	cache map[string]any,
	nodeId string) (*awsconfig.OrgNode, error) {

	var id, arn, name, nodeType string
	if strings.HasPrefix(nodeId, "r-") {
		nodeType = "ROOT"

		root, err := describeOrgRoot(ctx, client, cache)
		if err != nil {
			return nil, err
		}

		id = *root.Id
		arn = *root.Arn
		name = *root.Name
	} else if isAccount(nodeId) {
		nodeType = "ACCOUNT"

		summary, err := describeAccount(ctx, client, cache, nodeId)
		if err != nil {
			return nil, err
		}

		id = *summary.Account.Id
		arn = *summary.Account.Arn
		name = *summary.Account.Name
	} else {
		nodeType = "ORGANIZATIONAL_UNIT"

		summary, err := describeOu(ctx, client, cache, nodeId)
		if err != nil {
			return nil, err
		}

		id = *summary.OrganizationalUnit.Id
		arn = *summary.OrganizationalUnit.Arn
		name = *summary.OrganizationalUnit.Name
	}

	scps, err := listScpsForNode(ctx, client, cache, nodeId)
	if err != nil {
		return nil, err
	}

	rcps, err := listRcpsForNode(ctx, client, cache, nodeId)
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
	cache map[string]any,
	path []string) ([]awsconfig.OrgNode, error) {

	nodes := make([]awsconfig.OrgNode, len(path))
	for i, nodeId := range path {
		node, err := orgNode(ctx, client, cache, nodeId)
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
	cache map[string]any,
	orgId string,
	path []string,
	node string) (*awsconfig.Account, error) {

	nodes, err := orgNodes(ctx, client, cache, path)
	if err != nil {
		return nil, err
	}

	summary, err := describeAccount(ctx, client, cache, node)
	if err != nil {
		return nil, err
	}

	return &awsconfig.Account{
		ConfigItem: awsconfig.ConfigItem{
			Type:      "Yams::Organizations::Account",
			AccountId: node,
			Region:    "global",
			Arn:       *summary.Account.Arn,
		},
		Configuration: awsconfig.AccountConfiguration{
			Name:     *summary.Account.Name,
			OrgId:    orgId,
			OrgPaths: orgPaths(orgId, path),
			OrgNodes: nodes,
		},
	}, nil
}

func describeScps(
	ctx context.Context,
	client *organizations.Client,
	cache map[string]any) ([]awsconfig.SCP, error) {
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

		org, err := describeOrg(ctx, client, cache)
		if err != nil {
			return nil, err
		}

		s := awsconfig.SCP{
			ConfigItem: awsconfig.ConfigItem{
				Type:      "Yams::Organizations::ServiceControlPolicy",
				Name:      *rawPolicy.Policy.PolicySummary.Name,
				Arn:       *rawPolicy.Policy.PolicySummary.Arn,
				Region:    "global",
				AccountId: *org.MasterAccountId,
			},
			Configuration: awsconfig.SCPConfiguration{
				Document: awsconfig.EncodedPolicy(policyDocument),
			},
		}
		structured = append(structured, s)
	}

	return structured, nil
}

func describeRcps(
	ctx context.Context,
	client *organizations.Client,
	cache map[string]any) ([]awsconfig.SCP, error) {
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

		org, err := describeOrg(ctx, client, cache)
		if err != nil {
			return nil, err
		}

		s := awsconfig.SCP{
			ConfigItem: awsconfig.ConfigItem{
				Type:      "Yams::Organizations::ResourceControlPolicy",
				Name:      *rawPolicy.Policy.PolicySummary.Name,
				Arn:       *rawPolicy.Policy.PolicySummary.Arn,
				Region:    "global",
				AccountId: *org.MasterAccountId,
			},
			Configuration: awsconfig.SCPConfiguration{
				Document: awsconfig.EncodedPolicy(policyDocument),
			},
		}
		structured = append(structured, s)
	}

	return structured, nil
}
