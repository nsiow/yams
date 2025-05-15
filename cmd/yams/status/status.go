package status

import (
	"github.com/nsiow/yams/cmd/yams/cli"
)

// Logic for the "status" subcommand
func Run(opts *cli.Flags) {
	url := cli.ApiUrl(opts.Server, "status")
	cli.GetReq(url)
}
