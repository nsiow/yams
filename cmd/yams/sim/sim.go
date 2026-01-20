package sim

import (
	"fmt"
	"io"
	"os"

	json "github.com/bytedance/sonic"
	"github.com/nsiow/yams/cmd/yams/cli"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/entities"
	v1 "github.com/nsiow/yams/pkg/server/api/v1"
)

// Logic for the "sim" subcommand
func Run(opts *cli.Flags) {
	cli.RequireServer(opts.Server)

	havePrincipal := opts.Principal != ""
	haveAction := opts.Action != ""
	haveResource := opts.Resource != ""

	overlay, err := loadOverlays(opts.OverlayFiles)
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

func loadOverlays(files []string) (*v1.Overlay, error) {
	type overlayItem struct {
		Type string
	}

	overlay := v1.Overlay{}

	for _, fn := range files {
		file, err := os.Open(fn)
		if err != nil {
			return nil, fmt.Errorf("could not open overlay file '%s': %v", fn, err)
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("could not read overlay file '%s': %v", fn, err)
		}

		var item overlayItem
		err = json.Unmarshal(content, &item)
		if err != nil {
			return nil, fmt.Errorf("could not decode overlay file '%s': %v", fn, err)
		}
		if item.Type == "" {
			return nil, fmt.Errorf("could not decode overlay file '%s': missing field 'Type'", fn)
		}

		_, err = file.Seek(0, io.SeekStart) // reset to re-process
		if err != nil {
			return nil, fmt.Errorf("could not reset overlay file '%s': %v", fn, err)
		}

		switch item.Type {
		case "AWS::IAM::Role", "AWS::IAM::User":
			var principal entities.Principal
			err = json.Unmarshal(content, &principal)
			if err != nil {
				return nil, fmt.Errorf("could not decode principal from overlay file '%s': %v", fn, err)
			}
			overlay.Principals = append(overlay.Principals, principal)
		case "AWS::IAM::Group":
			var group entities.Group
			err = json.Unmarshal(content, &group)
			if err != nil {
				return nil, fmt.Errorf("could not decode group from overlay file '%s': %v", fn, err)
			}
			overlay.Groups = append(overlay.Groups, group)
		case
			"AWS::IAM::Policy",
			"Yams::Organizations::ServiceControlPolicy",
			"Yams::Organizations::ResourceControlPolicy":
			var policy entities.ManagedPolicy
			err = json.Unmarshal(content, &policy)
			if err != nil {
				return nil, fmt.Errorf("could not decode policy from overlay file '%s': %v", fn, err)
			}
			overlay.Policies = append(overlay.Policies, policy)
		case "Yams::Organizations::Account":
			var account entities.Account
			err = json.Unmarshal(content, &account)
			if err != nil {
				return nil, fmt.Errorf("could not decode account from overlay file '%s': %v", fn, err)
			}
			overlay.Accounts = append(overlay.Accounts, account)
		}

		// TODO(nsiow) figure out if we want to include accounts here or not; they aren't _really_
		// valid resources
		var resource entities.Resource
		err = json.Unmarshal(content, &resource)
		if err != nil {
			return nil, fmt.Errorf("could not decode resource from overlay file '%s': %v", fn, err)
		}
		overlay.Resources = append(overlay.Resources, resource)
	}

	return &overlay, nil
}
