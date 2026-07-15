package v1

import (
	"fmt"
	"net/http"

	json "github.com/bytedance/sonic"
	"github.com/nsiow/yams/pkg/server/httputil"
	"github.com/nsiow/yams/pkg/sim"
)

// -------------------------------------------------------------------------------------------------
// Schemas
// -------------------------------------------------------------------------------------------------

type WhichPrincipalsInput struct {
	Action               string              `json:"action"`
	Resource             string              `json:"resource"`
	Context              map[string]string   `json:"context"`
	MultiValueContext    map[string][]string `json:"multiValueContext"`
	DisableSharedContext bool                `json:"disableSharedContext"`

	Overlay Overlay `json:"overlay"`

	Fuzzy bool `json:"fuzzy"`
}

type WhichPrincipalsOutput = []string

// -------------------------------------------------------------------------------------------------
// Handlers
// -------------------------------------------------------------------------------------------------

func (api *API) WhichPrincipals(w http.ResponseWriter, req *http.Request) {
	input := WhichPrincipalsInput{}
	decoder := json.ConfigDefault.NewDecoder(req.Body)
	err := decoder.Decode(&input)
	if err != nil {
		httputil.ClientError(w, req, fmt.Errorf("invalid JSON: %v", err))
		return
	}

	if len(input.Action) == 0 {
		httputil.ClientError(w, req, fmt.Errorf("missing required field: action"))
		return
	}

	requestContext := api.requestContext(input.Context, input.DisableSharedContext)
	opts := sim.NewOptions(
		sim.WithAdditionalProperties(requestContext),
		sim.WithAdditionalMultiValueProperties(input.MultiValueContext),
	)
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
