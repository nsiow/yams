package server

import (
	"fmt"
	"net/http"

	"github.com/nsiow/yams/cmd/yams/cli"
	"github.com/nsiow/yams/internal/middleware"
	v1 "github.com/nsiow/yams/pkg/server/api/v1"
	"github.com/nsiow/yams/pkg/sim"
)

type Server struct {
	*http.Server
	mux *http.ServeMux

	Sources   []*Source
	Simulator *sim.Simulator
	Opts      *cli.Flags
}

func NewServer(opts *cli.Flags) (*Server, error) {
	mux := http.NewServeMux()
	server := Server{
		Server: &http.Server{
			Addr:    opts.Addr,
			Handler: middleware.Gzip(corsMiddleware(mux)),
		},
		mux:  mux,
		Opts: opts,
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

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}
