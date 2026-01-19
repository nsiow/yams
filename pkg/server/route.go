package server

import (
	v1 "github.com/nsiow/yams/pkg/server/api/v1"
)

func (s *Server) addV1Routes(v1 *v1.API) {
	// administration
	s.mux.HandleFunc("GET /api/v1/healthcheck", s.Healthcheck)
	s.mux.HandleFunc("GET /api/v1/status", s.Status)

	// accounts
	s.mux.HandleFunc("GET /api/v1/accounts", v1.ListAccounts)
	s.mux.HandleFunc("GET /api/v1/accounts/names", v1.AccountNames)
	s.mux.HandleFunc("GET /api/v1/accounts/{key...}", v1.GetAccount)
	s.mux.HandleFunc("GET /api/v1/accounts/search/{search...}", v1.SearchAccounts)

	// groups
	s.mux.HandleFunc("GET /api/v1/groups", v1.ListGroups)
	s.mux.HandleFunc("GET /api/v1/groups/{key...}", v1.GetGroup)
	s.mux.HandleFunc("GET /api/v1/groups/search/{search...}", v1.SearchGroups)

	// policies
	s.mux.HandleFunc("GET /api/v1/policies", v1.ListPolicies)
	s.mux.HandleFunc("GET /api/v1/policies/{key...}", v1.GetPolicy)
	s.mux.HandleFunc("GET /api/v1/policies/search/{search...}", v1.SearchPolicies)

	// principals
	s.mux.HandleFunc("GET /api/v1/principals", v1.ListPrincipals)
	s.mux.HandleFunc("GET /api/v1/principals/{key...}", v1.GetPrincipal)
	s.mux.HandleFunc("GET /api/v1/principals/search/{search...}", v1.SearchPrincipals)

	// resources
	s.mux.HandleFunc("GET /api/v1/resources", v1.ListResources)
	s.mux.HandleFunc("GET /api/v1/resources/{key...}", v1.GetResource)
	s.mux.HandleFunc("GET /api/v1/resources/search/{search...}", v1.SearchResources)

	// actions
	s.mux.HandleFunc("GET /api/v1/actions", v1.ListActions)
	s.mux.HandleFunc("GET /api/v1/actions/{key...}", v1.GetAction)
	s.mux.HandleFunc("GET /api/v1/actions/search/{search...}", v1.SearchActions)

	// simulation
	s.mux.HandleFunc("POST /api/v1/sim", v1.SimRun)
	s.mux.HandleFunc("POST /api/v1/sim/whichPrincipals", v1.WhichPrincipals)
	s.mux.HandleFunc("POST /api/v1/sim/whichActions", v1.WhichActions)
	s.mux.HandleFunc("POST /api/v1/sim/whichResources", v1.WhichResources)

	// utils
	s.mux.HandleFunc("GET /api/v1/utils/accounts/names", v1.UtilAccountNames)
	s.mux.HandleFunc("GET /api/v1/utils/resources/accounts", v1.UtilResourceAccounts)
	s.mux.HandleFunc("GET /api/v1/utils/actions/resourceless", v1.UtilResourcelessActions)
	s.mux.HandleFunc("GET /api/v1/utils/actions/accesslevels", v1.UtilActionAccessLevels)
}
