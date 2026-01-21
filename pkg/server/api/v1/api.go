package v1

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/nsiow/yams/internal/common"
	"github.com/nsiow/yams/pkg/aws/sar"
	"github.com/nsiow/yams/pkg/aws/sar/types"
	"github.com/nsiow/yams/pkg/server/httputil"
	"github.com/nsiow/yams/pkg/sim"
)

// -------------------------------------------------------------------------------------------------
// API
// -------------------------------------------------------------------------------------------------

type API struct {
	Simulator *sim.Simulator
}

// -------------------------------------------------------------------------------------------------
// Get
// -------------------------------------------------------------------------------------------------

func (api *API) GetAccount(w http.ResponseWriter, req *http.Request) {
	Get(w, req, api.Simulator.Universe.Account)
}

func (api *API) GetGroup(w http.ResponseWriter, req *http.Request) {
	Get(w, req, api.Simulator.Universe.Group)
}

func (api *API) GetPolicy(w http.ResponseWriter, req *http.Request) {
	Get(w, req, api.Simulator.Universe.Policy)
}

func (api *API) GetPrincipal(w http.ResponseWriter, req *http.Request) {
	Get(w, req, api.Simulator.Universe.Principal)
}

func (api *API) GetResource(w http.ResponseWriter, req *http.Request) {
	Get(w, req, api.Simulator.Universe.Resource)
}

// -------------------------------------------------------------------------------------------------
// List
// -------------------------------------------------------------------------------------------------

func (api *API) ListAccounts(w http.ResponseWriter, req *http.Request) {
	List(w, req, api.Simulator.Universe.Accounts)
}

func (api *API) AccountNames(w http.ResponseWriter, req *http.Request) {
	names := make(map[string]string)
	for account := range api.Simulator.Universe.Accounts() {
		names[account.Id] = account.Name
	}
	httputil.WriteJsonResponse(w, req, names)
}

func (api *API) ListGroups(w http.ResponseWriter, req *http.Request) {
	List(w, req, api.Simulator.Universe.Groups)
}

func (api *API) ListPolicies(w http.ResponseWriter, req *http.Request) {
	List(w, req, api.Simulator.Universe.Policies)
}

func (api *API) ListPrincipals(w http.ResponseWriter, req *http.Request) {
	List(w, req, api.Simulator.Universe.Principals)
}

func (api *API) ListResources(w http.ResponseWriter, req *http.Request) {
	keys := []string{}
	for entity := range api.Simulator.Universe.Resources() {
		keys = append(keys, entity.Key())
	}
	keys = expandS3Buckets(keys)
	slog.Debug("serving up resources", "count", len(keys))
	slices.Sort(keys)
	httputil.WriteJsonResponse(w, req, keys)
}

// -------------------------------------------------------------------------------------------------
// Search
// -------------------------------------------------------------------------------------------------

func (api *API) SearchAccounts(w http.ResponseWriter, req *http.Request) {
	Search(w, req, api.Simulator.Universe.Accounts)
}

func (api *API) SearchGroups(w http.ResponseWriter, req *http.Request) {
	Search(w, req, api.Simulator.Universe.Groups)
}

func (api *API) SearchPolicies(w http.ResponseWriter, req *http.Request) {
	Search(w, req, api.Simulator.Universe.Policies)
}

func (api *API) SearchPrincipals(w http.ResponseWriter, req *http.Request) {
	Search(w, req, api.Simulator.Universe.Principals)
}

func (api *API) SearchResources(w http.ResponseWriter, req *http.Request) {
	search := strings.ToLower(req.PathValue("search"))
	actionFilter := req.URL.Query().Get("action")

	// Look up action for filtering if specified
	var action *types.Action
	if actionFilter != "" {
		if a, ok := sar.LookupString(actionFilter); ok {
			action = a
		}
	}

	keys := []string{}
	for entity := range api.Simulator.Universe.Resources() {
		key := entity.Key()
		if strings.Contains(strings.ToLower(key), search) {
			keys = append(keys, key)
		}
	}
	keys = expandS3Buckets(keys)

	// Filter by action targeting if action was specified
	if action != nil {
		filtered := []string{}
		for _, key := range keys {
			if action.Targets(key) {
				filtered = append(filtered, key)
			}
		}
		keys = filtered
	}

	slog.Debug("serving up resources", "count", len(keys))
	slices.Sort(keys)
	httputil.WriteJsonResponse(w, req, keys)
}

// -------------------------------------------------------------------------------------------------
// SAR
// -------------------------------------------------------------------------------------------------

func (api *API) ListActions(w http.ResponseWriter, req *http.Request) {
	api.SearchActions(w, req) // search with empty string is equivalent of list
}

func (api *API) GetAction(w http.ResponseWriter, req *http.Request) {
	key := req.PathValue("key")
	if len(key) == 0 {
		httputil.ClientError(w, req, fmt.Errorf("no action provided"))
		return
	}

	action, ok := sar.LookupString(key)
	if !ok {
		httputil.Error(w, req, http.StatusNotFound, fmt.Errorf("unknown action: '%s'", key))
		return
	}

	httputil.WriteJsonResponse(w, req, action)
}

func (api *API) SearchActions(w http.ResponseWriter, req *http.Request) {
	search := req.PathValue("search")
	results := sar.NewQuery().WithSearch(search).Results()
	actions := common.Map(results, func(in types.Action) string { return in.ShortName() })

	httputil.WriteJsonResponse(w, req, actions)
}

// expandS3Buckets takes a list of resource ARNs and expands S3 bucket ARNs to include
// a synthetic S3 object ARN. This enables simulating S3 object operations on buckets.
func expandS3Buckets(keys []string) []string {
	expanded := make([]string, 0, len(keys)*2)
	for _, key := range keys {
		expanded = append(expanded, key)
		// S3 bucket ARNs: arn:aws:s3:::bucket-name (no slash)
		// S3 object ARNs: arn:aws:s3:::bucket-name/key (has slash)
		if strings.HasPrefix(key, "arn:aws:s3:::") && !strings.Contains(key, "/") {
			expanded = append(expanded, key+"/object.txt")
		}
	}
	return expanded
}
