package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/nsiow/yams/pkg/loaders/awsconfig"
)

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

	return *resp.Roots[0].Id, nil
}

func orgPaths(orgId string, path []string) []string {
	var paths []string

	segment := orgId + "/"

	for _, p := range path {
		segment += p + "/"
		paths = append(paths, segment)
	}

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
	_walk(ctx, client, id, accounts, []string{root})

	return accounts, nil
}

func fetchSCPs(
	ctx context.Context,
	client *organizations.Client,
	path []string,
	node string) ([][]string, error) {
	return nil, nil
}

func fetchRCPs(
	ctx context.Context,
	client *organizations.Client,
	path []string,
	node string) ([][]string, error) {
	return nil, nil
}

func _account(
	ctx context.Context,
	client *organizations.Client,
	orgId string,
	path []string,
	node string) (*awsconfig.Account, error) {
	scps, err := fetchSCPs(ctx, client, path, node)
	if err != nil {
		return nil, err
	}

	rcps, err := fetchRCPs(ctx, client, path, node)
	if err != nil {
		return nil, err
	}

	return &awsconfig.Account{
		ConfigItem: awsconfig.ConfigItem{
			AccountId: node,
		},
		Configuration: awsconfig.AccountConfiguration{
			OrgId:    orgId,
			OrgPaths: orgPaths(orgId, path),
			SCPs:     scps,
			RCPs:     rcps,
		},
	}, nil
}

func _walk(
	ctx context.Context,
	client *organizations.Client,
	orgId string,
	accounts []awsconfig.Account, path []string) error {
	node := path[len(path)-1]

	if isAccount(node) {
		a, err := _account(ctx, client, orgId, path, node)
		if err != nil {
			return err
		}

		accounts = append(accounts, *a)
	}

	return nil
}
