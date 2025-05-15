package dump

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/nsiow/yams/cmd/yams/cli"
	"github.com/nsiow/yams/internal/common"
)

// -------------------------------------------------------------------------------------------------
// Dumping logic
// -------------------------------------------------------------------------------------------------

func Config(opts *cli.Flags) (string, error) {
	client, err := configClient(context.TODO())
	if err != nil {
		return "", err
	}

	query, err := buildQuery(opts.ResourceTypes)
	if err != nil {
		return "", err
	}

	raw, err := executeQuery(client, query, opts.Aggregator)
	if err != nil {
		return "", err
	}

	return parseResults(raw)
}

// parseResults converts the raw output of select-aggregate-resource-config into valid JSON
func parseResults(raw []string) (string, error) {
	var parsed []map[string]any

	for _, blob := range raw {
		var m map[string]any
		err := json.Unmarshal([]byte(blob), &m)
		if err != nil {
			return "", fmt.Errorf("unable to parse config blob (%w): %s", err, blob)
		}
		parsed = append(parsed, m)
	}

	asJson, err := json.Marshal(parsed)
	return string(asJson), err
}

// executeQuery runs the provided query using the
func executeQuery(client *configservice.Client, query string, aggregator string) ([]string, error) {
	if len(aggregator) == 0 {
		return nil, fmt.Errorf("must provide aggregator name via '-a' flag")
	}

	var results []string
	var nextToken *string

	for {
		resp, err := client.SelectAggregateResourceConfig(
			context.TODO(),
			&configservice.SelectAggregateResourceConfigInput{
				ConfigurationAggregatorName: &aggregator,
				Expression:                  &query,
				NextToken:                   nextToken,
				MaxResults:                  100, // maximum
			},
		)
		if err != nil {
			return nil, fmt.Errorf("error performing select-aggregate-resource-config: %w", err)
		}

		results = append(results, resp.Results...)

		if resp.NextToken == nil {
			break
		}
		nextToken = resp.NextToken
	}

	return results, nil
}

// buildQuery constructs an AWS Config Advanced Query based on the provided parameters
func buildQuery(rtypes []string) (string, error) {
	if len(rtypes) == 0 {
		return "", fmt.Errorf("must provide one or more resource types via '-r' flag")
	}

	for _, rtype := range rtypes {
		if !strings.HasPrefix(rtype, "AWS::") {
			return "", fmt.Errorf("invalid resource type: %s", rtype)
		}
	}

	template := `
		SELECT
			*,
			configuration,
			supplementaryConfiguration,
			tags
		WHERE
			resourceType IN (%s)
	`

	quoted := common.Map(rtypes, func(x string) string { return "'" + x + "'" })
	joined := strings.Join(quoted, ",")
	return fmt.Sprintf(template, joined), nil
}

// configClient creates and returns a new AWS Config SDK client using the provided options
func configClient(ctx context.Context) (*configservice.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating config client: %w", err)
	}

	client := configservice.NewFromConfig(cfg)
	return client, nil
}
