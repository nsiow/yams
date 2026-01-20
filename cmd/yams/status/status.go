package status

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	json "github.com/bytedance/sonic"
	"github.com/nsiow/yams/cmd/yams/cli"
)

// statusResponse represents the server status response
type statusResponse struct {
	Server  serverInfo  `json:"server"`
	Sources []sourceInfo `json:"sources"`
}

type serverInfo struct {
	Started string            `json:"started"`
	Env     map[string]string `json:"env"`
}

type sourceInfo struct {
	Name       string `json:"name"`
	Principals int    `json:"principals"`
	Resources  int    `json:"resources"`
	Policies   int    `json:"policies"`
	Actions    int    `json:"actions"`
	Accounts   int    `json:"accounts"`
}

// Logic for the "status" subcommand
func Run(opts *cli.Flags) {
	url := cli.ApiUrl(opts.Server, "status")

	if opts.Format == cli.FormatTable {
		getAndRenderStatus(url)
	} else {
		cli.GetReq(url)
	}
}

func getAndRenderStatus(url string) {
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

	var status statusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		cli.OutputJSON(body)
		return
	}

	// Print server info
	fmt.Fprintln(os.Stdout, "Server Status")
	fmt.Fprintln(os.Stdout, "-------------")
	fmt.Fprintf(os.Stdout, "Started: %s\n", status.Server.Started)

	if len(status.Server.Env) > 0 {
		fmt.Fprintln(os.Stdout, "\nEnvironment:")
		for k, v := range status.Server.Env {
			fmt.Fprintf(os.Stdout, "  %s: %s\n", k, v)
		}
	}

	// Print sources
	if len(status.Sources) > 0 {
		fmt.Fprintln(os.Stdout, "\nData Sources:")
		t := cli.NewTableWriter("Source", "Principals", "Resources", "Policies", "Actions", "Accounts")
		for _, src := range status.Sources {
			t.AddRow(
				src.Name,
				fmt.Sprintf("%d", src.Principals),
				fmt.Sprintf("%d", src.Resources),
				fmt.Sprintf("%d", src.Policies),
				fmt.Sprintf("%d", src.Actions),
				fmt.Sprintf("%d", src.Accounts),
			)
		}
		t.Render()
	}
}
