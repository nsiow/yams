package v1

import (
	"fmt"
	"net/http"

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
	List(w, req, api.Simulator.Universe.Resources)
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
	Search(w, req, api.Simulator.Universe.Resources)
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
