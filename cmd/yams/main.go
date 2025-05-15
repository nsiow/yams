package main

import (
	"log/slog"

	"github.com/nsiow/yams/cmd/yams/cli"
	"github.com/nsiow/yams/cmd/yams/dump"
	"github.com/nsiow/yams/cmd/yams/inventory"
	"github.com/nsiow/yams/cmd/yams/server"
	"github.com/nsiow/yams/cmd/yams/sim"
	"github.com/nsiow/yams/cmd/yams/status"
)

func main() {
	// Set up CLI logging
	cli.InitLogging()

	// Parse CLI arguments
	flags, err := cli.Parse()
	if err != nil {
		cli.Fail("%s", err.Error())
	}
	slog.Debug("cli flags", "flags", flags)

	// Run the requested command logic
	switch flags.Mode {
	case cli.RUN_MODE_STATUS:
		status.Run(flags)
	case cli.RUN_MODE_DUMP:
		dump.Run(flags)
	case
		cli.RUN_MODE_ACCOUNTS,
		cli.RUN_MODE_ACTIONS,
		cli.RUN_MODE_POLICIES,
		cli.RUN_MODE_PRINCIPALS,
		cli.RUN_MODE_RESOURCES:
		inventory.Run(flags.Mode, flags)
	case cli.RUN_MODE_SERVER:
		server.Run(flags)
	case cli.RUN_MODE_SIM:
		sim.Run(flags)
	default:
		cli.Fail("unknown mode: %s", flags.Mode)
	}
}
