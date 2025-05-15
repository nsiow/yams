package server

import (
	"fmt"
	"net/http"

	v1 "github.com/nsiow/yams/pkg/server/api/v1"
	"github.com/nsiow/yams/pkg/sim"
)

type Server struct {
	*http.Server
	mux *http.ServeMux

	Sources   []*Source
	Simulator *sim.Simulator
}

func NewServer(addr string) (*Server, error) {
	mux := http.NewServeMux()
	server := Server{
		Server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
		mux: mux,
	}

	sim, err := sim.NewSimulator()
	if err != nil {
		return nil, fmt.Errorf("unable create simulator: %w", err)
	}
	server.Simulator = sim

	// routes routes routes
	server.addV1Routes(&v1.API{Simulator: server.Simulator})

	return &server, nil
}
