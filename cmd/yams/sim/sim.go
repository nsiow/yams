package sim

import (
	"github.com/nsiow/yams/cmd/yams/cli"
	"github.com/nsiow/yams/pkg/aws/sar"
	v1 "github.com/nsiow/yams/pkg/server/api/v1"
)

// Logic for the "sim" subcommand
func Run(opts *cli.Flags) {
	cli.RequireServer(opts.Server)

	havePrincipal := opts.Principal != ""
	haveAction := opts.Action != ""
	haveResource := opts.Resource != ""

	overlay, err := cli.LoadOverlays(opts.OverlayFiles)
	if err != nil {
		cli.Fail("error loading overlays: %v", err)
	}
	opts.Overlay = *overlay

	if havePrincipal && haveAction && haveResource {
		runSim(opts)
	} else if havePrincipal && haveAction {
		action, ok := sar.LookupString(opts.Action)
		if !ok {
			cli.Fail("unknown action: %s", opts.Action)
		}

		if action.HasTargets() {
			runWhichResources(opts)
		} else {
			runSim(opts)
		}
	} else if havePrincipal && haveResource {
		runWhichActions(opts)
	} else if haveAction && haveResource {
		runWhichPrincipals(opts)
	} else {
		cli.Fail("error: must provide at least two of -p/--principal | -a/--action | -r/--resource")
	}
}

func runSim(opts *cli.Flags) {
	cli.PostReq(
		cli.ApiUrl(opts.Server, "sim"),
		v1.SimInput{
			Principal: opts.Principal,
			Action:    opts.Action,
			Resource:  opts.Resource,
			Context:   opts.Context,
			Fuzzy:     !opts.Exact,
			Explain:   opts.Explain,
			Trace:     opts.Trace,
			Overlay:   opts.Overlay,
		},
	)
}

func runWhichPrincipals(opts *cli.Flags) {
	cli.PostReq(
		cli.ApiUrl(opts.Server, "sim", "whichPrincipals"),
		v1.WhichPrincipalsInput{
			Action:   opts.Action,
			Resource: opts.Resource,
			Context:  opts.Context,
			Overlay:  opts.Overlay,
			Fuzzy:    !opts.Exact,
		},
	)
}

func runWhichActions(opts *cli.Flags) {
	cli.PostReq(
		cli.ApiUrl(opts.Server, "sim", "whichActions"),
		v1.WhichActionsInput{
			Principal: opts.Principal,
			Resource:  opts.Resource,
			Context:   opts.Context,
			Overlay:   opts.Overlay,
			Fuzzy:     !opts.Exact,
		},
	)
}

func runWhichResources(opts *cli.Flags) {
	cli.PostReq(
		cli.ApiUrl(opts.Server, "sim", "whichResources"),
		v1.WhichResourcesInput{
			Principal: opts.Principal,
			Action:    opts.Action,
			Context:   opts.Context,
			Overlay:   opts.Overlay,
			Fuzzy:     !opts.Exact,
		},
	)
}

