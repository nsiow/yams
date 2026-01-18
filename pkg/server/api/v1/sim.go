package v1

import (
	"fmt"
	"log/slog"
	"net/http"

	json "github.com/bytedance/sonic"
	"github.com/nsiow/yams/pkg/server/httputil"
	"github.com/nsiow/yams/pkg/sim"
)

// -------------------------------------------------------------------------------------------------
// Schemas
// -------------------------------------------------------------------------------------------------

type SimInput struct {
	Principal string            `json:"principal"`
	Action    string            `json:"action"`
	Resource  string            `json:"resource"`
	Context   map[string]string `json:"context"`

	Fuzzy   bool    `json:"fuzzy"`
	Explain bool    `json:"explain"`
	Trace   bool    `json:"trace"`
	Overlay Overlay `json:"overlay"`
}

type SimOutput struct {
	Result    string   `json:"result"`
	Principal string   `json:"principal"`
	Action    string   `json:"action"`
	Resource  string   `json:"resource,omitzero"`
	Explain   []string `json:"explain,omitzero"`
	Trace     []string `json:"trace,omitzero"`
}

// -------------------------------------------------------------------------------------------------
// Handlers
// -------------------------------------------------------------------------------------------------

func (api *API) SimRun(w http.ResponseWriter, req *http.Request) {
	// read input
	input := SimInput{}
	decoder := json.ConfigDefault.NewDecoder(req.Body)
	err := decoder.Decode(&input)
	if err != nil {
		httputil.ClientError(w, req, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	// validate
	if len(input.Principal) == 0 {
		httputil.ClientError(w, req, fmt.Errorf("missing required input 'principal'"))
		return
	}
	if len(input.Action) == 0 {
		httputil.ClientError(w, req, fmt.Errorf("missing required input 'action'"))
		return
	}

	// construct options
	opts := sim.NewOptions(sim.WithAdditionalProperties(input.Context))
	opts.EnableTracing = input.Explain || input.Trace
	opts.Overlay = input.Overlay.Universe()
	opts.EnableFuzzyMatchArn = input.Fuzzy

	// simulate
	result, err := api.Simulator.SimulateByArnWithOptions(
		input.Principal,
		input.Action,
		input.Resource,
		opts)
	if err != nil {
		httputil.ServerError(w, req, fmt.Errorf("simulation error: %v", err))
		return
	}

	// construct response
	out := SimOutput{}
	out.Principal = result.Principal
	out.Action = result.Action
	out.Resource = result.Resource
	if result.IsAllowed {
		out.Result = "ALLOW"
	} else {
		out.Result = "DENY"
	}
	if input.Explain || input.Trace {
		out.Explain = result.Trace.Explain()
	}
	if input.Trace {
		out.Trace = result.Trace.Trace()
	}

	slog.Info("simulation result",
		"principal", input.Principal,
		"action", input.Action,
		"resource", input.Resource,
		"context", input.Context,
		"result", out.Result)
	httputil.WriteJsonResponse(w, req, out)
}
