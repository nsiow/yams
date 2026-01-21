package server

import (
	v1 "github.com/nsiow/yams/pkg/server/api/v1"
)

func (s *Server) addV1Routes(api *v1.API, overlayAPI *v1.OverlayAPI) {
	// administration
	s.mux.HandleFunc("GET /api/v1/healthcheck", s.Healthcheck)
	s.mux.HandleFunc("GET /api/v1/status", s.Status)

	// accounts
	s.mux.HandleFunc("GET /api/v1/accounts", api.ListAccounts)
	s.mux.HandleFunc("GET /api/v1/accounts/names", api.AccountNames)
	s.mux.HandleFunc("GET /api/v1/accounts/{key...}", api.GetAccount)
	s.mux.HandleFunc("GET /api/v1/accounts/search/{search...}", api.SearchAccounts)

	// groups
	s.mux.HandleFunc("GET /api/v1/groups", api.ListGroups)
	s.mux.HandleFunc("GET /api/v1/groups/{key...}", api.GetGroup)
	s.mux.HandleFunc("GET /api/v1/groups/search/{search...}", api.SearchGroups)

	// policies
	s.mux.HandleFunc("GET /api/v1/policies", api.ListPolicies)
	s.mux.HandleFunc("GET /api/v1/policies/{key...}", api.GetPolicy)
	s.mux.HandleFunc("GET /api/v1/policies/search/{search...}", api.SearchPolicies)

	// principals
	s.mux.HandleFunc("GET /api/v1/principals", api.ListPrincipals)
	s.mux.HandleFunc("GET /api/v1/principals/{key...}", api.GetPrincipal)
	s.mux.HandleFunc("GET /api/v1/principals/search/{search...}", api.SearchPrincipals)

	// resources
	s.mux.HandleFunc("GET /api/v1/resources", api.ListResources)
	s.mux.HandleFunc("GET /api/v1/resources/{key...}", api.GetResource)
	s.mux.HandleFunc("GET /api/v1/resources/search/{search...}", api.SearchResources)

	// actions
	s.mux.HandleFunc("GET /api/v1/actions", api.ListActions)
	s.mux.HandleFunc("GET /api/v1/actions/{key...}", api.GetAction)
	s.mux.HandleFunc("GET /api/v1/actions/search/{search...}", api.SearchActions)

	// simulation
	s.mux.HandleFunc("POST /api/v1/sim", api.SimRun)
	s.mux.HandleFunc("POST /api/v1/sim/whichPrincipals", api.WhichPrincipals)
	s.mux.HandleFunc("POST /api/v1/sim/whichActions", api.WhichActions)
	s.mux.HandleFunc("POST /api/v1/sim/whichResources", api.WhichResources)

	// utils
	s.mux.HandleFunc("GET /api/v1/utils/accounts/names", api.UtilAccountNames)
	s.mux.HandleFunc("GET /api/v1/utils/resources/accounts", api.UtilResourceAccounts)
	s.mux.HandleFunc("GET /api/v1/utils/actions/resourceless", api.UtilResourcelessActions)
	s.mux.HandleFunc("GET /api/v1/utils/actions/accesslevels", api.UtilActionAccessLevels)
	s.mux.HandleFunc("GET /api/v1/utils/actions/targeting", api.UtilActionTargeting)

	// overlays
	s.mux.HandleFunc("GET /api/v1/overlays", overlayAPI.ListOverlays)
	s.mux.HandleFunc("POST /api/v1/overlays", overlayAPI.CreateOverlay)
	s.mux.HandleFunc("GET /api/v1/overlays/{id}", overlayAPI.GetOverlay)
	s.mux.HandleFunc("PUT /api/v1/overlays/{id}", overlayAPI.UpdateOverlay)
	s.mux.HandleFunc("DELETE /api/v1/overlays/{id}", overlayAPI.DeleteOverlay)
}
