package inventory

import (
	"io"
	"log/slog"
	"net/http"

	json "github.com/bytedance/sonic"
	"github.com/nsiow/yams/cmd/yams/cli"
)

// inventoryItem represents a generic inventory entity for table display
type inventoryItem struct {
	Arn       string `json:"Arn"`
	Name      string `json:"Name"`
	Type      string `json:"Type"`
	AccountId string `json:"AccountId"`
	Region    string `json:"Region"`
}

// Logic for the various entity-centric subcommands (accounts, resources, principals, etc)
func Run(entity string, opts *cli.Flags) {
	cli.RequireServer(opts.Server)

	var url string
	if opts.Key != "" {
		url = cli.ApiUrl(opts.Server, entity, opts.Key)
	} else if opts.Query != "" {
		url = cli.ApiUrl(opts.Server, entity, "search", opts.Query)
	} else {
		url = cli.ApiUrl(opts.Server, entity)
	}

	if opts.Freeze && entity != "actions" {
		url += "/freeze"
	}

	if opts.Format == cli.FormatTable {
		getAndRenderTable(url, entity)
	} else {
		cli.GetReq(url)
	}
}

func getAndRenderTable(url, entity string) {
	slog.Debug("making request", "url", url)
	resp, err := http.Get(url)
	if err != nil {
		cli.Fail("error retrieving URL '%s': %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		cli.Fail("received non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		cli.Fail("error reading body of URL '%s': %v", url, err)
	}

	// Try to parse as array
	var items []inventoryItem
	if err := json.Unmarshal(body, &items); err != nil {
		// Fall back to JSON output if parsing fails
		cli.OutputJSON(body)
		return
	}

	// Render as table based on entity type
	switch entity {
	case "principals":
		renderPrincipalsTable(items)
	case "resources":
		renderResourcesTable(items)
	case "accounts":
		renderAccountsTable(items)
	case "policies":
		renderPoliciesTable(items)
	case "actions":
		renderActionsTable(body)
	default:
		cli.OutputJSON(body)
	}
}

func renderPrincipalsTable(items []inventoryItem) {
	t := cli.NewTableWriter("Type", "Name", "Account", "Arn")
	for _, item := range items {
		t.AddRow(item.Type, item.Name, item.AccountId, cli.Truncate(item.Arn, 60))
	}
	t.Render()
}

func renderResourcesTable(items []inventoryItem) {
	t := cli.NewTableWriter("Type", "Name", "Account", "Region", "Arn")
	for _, item := range items {
		t.AddRow(item.Type, item.Name, item.AccountId, item.Region, cli.Truncate(item.Arn, 50))
	}
	t.Render()
}

func renderAccountsTable(items []inventoryItem) {
	t := cli.NewTableWriter("AccountId", "Name", "Arn")
	for _, item := range items {
		t.AddRow(item.AccountId, item.Name, item.Arn)
	}
	t.Render()
}

func renderPoliciesTable(items []inventoryItem) {
	t := cli.NewTableWriter("Type", "Name", "Account", "Arn")
	for _, item := range items {
		t.AddRow(item.Type, item.Name, item.AccountId, cli.Truncate(item.Arn, 60))
	}
	t.Render()
}

func renderActionsTable(body []byte) {
	var actions []string
	if err := json.Unmarshal(body, &actions); err != nil {
		cli.OutputJSON(body)
		return
	}

	t := cli.NewTableWriter("Action")
	for _, action := range actions {
		t.AddRow(action)
	}
	t.Render()
}
