package server

import (
	"net/http"

	"github.com/nsiow/yams/pkg/server/httputil"
)

func (s *Server) Healthcheck(w http.ResponseWriter, req *http.Request) {
	httputil.WriteJsonResponse(w, req, map[string]string{"status": "ok"})
}
