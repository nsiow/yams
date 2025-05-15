package inventory

import (
	"github.com/nsiow/yams/cmd/yams/cli"
)

// Logic for the various entity-centric subcommands (accounts, resources, principals, etc)
func Run(entity string, opts *cli.Flags) {
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

	cli.GetReq(url)
}
