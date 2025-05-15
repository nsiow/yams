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

type WhichActionsInput struct {
	Principal string            `json:"principal"`
	Resource  string            `json:"resource"`
	Context   map[string]string `json:"context"`

	Overlay Overlay `json:"overlay"`

	Fuzzy bool `json:"fuzzy"`
}

type WhichActionsOutput = []string

// -------------------------------------------------------------------------------------------------
// Handlers
// -------------------------------------------------------------------------------------------------

func (api *API) WhichActions(w http.ResponseWriter, req *http.Request) {
	input := WhichActionsInput{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&input)
	if err != nil {
		httputil.ClientError(w, req, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	opts := sim.NewOptions(sim.WithAdditionalProperties(input.Context))
	opts.Overlay = input.Overlay.Universe()
	opts.EnableFuzzyMatchArn = input.Fuzzy

	resources, err := api.Simulator.WhichActions(input.Principal, input.Resource, opts)
	if err != nil {
		httputil.ServerError(w, req, fmt.Errorf("simulation error: %v", err))
		return
	}

	var out []string = resources
	httputil.WriteJsonResponse(w, req, out)
}
