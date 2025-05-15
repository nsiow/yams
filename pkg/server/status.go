package server

import (
	"net/http"

	"github.com/nsiow/yams/internal/common"
	"github.com/nsiow/yams/pkg/server/httputil"
)

func (s *Server) Status(w http.ResponseWriter, req *http.Request) {
	status := map[string]any{
		"entities":   s.Simulator.Universe.Size(),
		"accounts":   s.Simulator.Universe.NumAccounts(),
		"principals": s.Simulator.Universe.NumPrincipals(),
		"groups":     s.Simulator.Universe.NumGroups(),
		"policies":   s.Simulator.Universe.NumPolicies(),
		"resources":  s.Simulator.Universe.NumResources(),
		"sources": common.Map(s.Sources, func(src *Source) map[string]any {
			return map[string]any{
				"source":  src.Reader.Source,
				"updated": src.Updated,
			}
		}),
	}

	httputil.WriteJsonResponse(w, req, status)
}
