package server

import (
	"fmt"
	"net/http"
)

func (s *Server) Healthcheck(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "OK\n")
}
