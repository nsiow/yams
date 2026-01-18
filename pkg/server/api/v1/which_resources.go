package v1

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nsiow/yams/pkg/server/httputil"
	"github.com/nsiow/yams/pkg/sim"
)

// -------------------------------------------------------------------------------------------------
// Schemas
// -------------------------------------------------------------------------------------------------

type WhichResourcesInput struct {
	Principal string            `json:"principal"`
	Action    string            `json:"action"`
	Context   map[string]string `json:"context"`

	Overlay Overlay `json:"overlay"`

	Fuzzy bool `json:"fuzzy"`
}

type WhichResourcesOutput = []string

// -------------------------------------------------------------------------------------------------
// Handlers
// -------------------------------------------------------------------------------------------------

func (api *API) WhichResources(w http.ResponseWriter, req *http.Request) {
	input := WhichResourcesInput{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&input)
	if err != nil {
		httputil.ClientError(w, req, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	if len(input.Principal) == 0 {
		httputil.ClientError(w, req, fmt.Errorf("missing required field: principal"))
		return
	}

	opts := sim.NewOptions(sim.WithAdditionalProperties(input.Context))
	opts.Overlay = input.Overlay.Universe()
	opts.EnableFuzzyMatchArn = input.Fuzzy

	resources, err := api.Simulator.WhichResources(input.Principal, input.Action, opts)
	if err != nil {
		httputil.ServerError(w, req, fmt.Errorf("simulation error: %v", err))
		return
	}

	var out []string = resources
	httputil.WriteJsonResponse(w, req, out)
}
