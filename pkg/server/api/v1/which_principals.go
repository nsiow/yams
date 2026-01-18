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

type WhichPrincipalsInput struct {
	Action   string            `json:"action"`
	Resource string            `json:"resource"`
	Context  map[string]string `json:"context"`

	Overlay Overlay `json:"overlay"`

	Fuzzy bool `json:"fuzzy"`
}

type WhichPrincipalsOutput = []string

// -------------------------------------------------------------------------------------------------
// Handlers
// -------------------------------------------------------------------------------------------------

func (api *API) WhichPrincipals(w http.ResponseWriter, req *http.Request) {
	input := WhichPrincipalsInput{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&input)
	if err != nil {
		httputil.ClientError(w, req, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	if len(input.Action) == 0 {
		httputil.ClientError(w, req, fmt.Errorf("missing required field: action"))
		return
	}

	opts := sim.NewOptions(sim.WithAdditionalProperties(input.Context))
	opts.Overlay = input.Overlay.Universe()
	opts.EnableFuzzyMatchArn = input.Fuzzy

	principals, err := api.Simulator.WhichPrincipals(input.Action, input.Resource, opts)
	if err != nil {
		httputil.ServerError(w, req, fmt.Errorf("simulation error: %v", err))
		return
	}

	var out []string = principals
	httputil.WriteJsonResponse(w, req, out)
}
