package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/nsiow/yams/cmd/yams/cli"
	"github.com/nsiow/yams/internal/middleware"
	"github.com/nsiow/yams/pkg/overlay"
	v1 "github.com/nsiow/yams/pkg/server/api/v1"
	"github.com/nsiow/yams/pkg/sim"
)

type Server struct {
	*http.Server
	mux *http.ServeMux

	Sources      []*Source
	Simulator    *sim.Simulator
	OverlayStore overlay.Store
	Opts         *cli.Flags
}

func NewServer(opts *cli.Flags) (*Server, error) {
	mux := http.NewServeMux()

	// Middleware chain: Cache -> Gzip -> CORS -> Handler
	handler := corsMiddleware(mux)
	handler = middleware.Gzip(handler)
	handler = middleware.Cache(5 * time.Minute)(handler)

	server := Server{
		Server: &http.Server{
			Addr:    opts.Addr,
			Handler: handler,
		},
		mux:  mux,
		Opts: opts,
	}

	sim, err := sim.NewSimulator()
	if err != nil {
		return nil, fmt.Errorf("unable create simulator: %w", err)
	}
	server.Simulator = sim

	// Create overlay store
	overlayStore, err := overlay.NewStore(opts.OverlayStore)
	if err != nil {
		return nil, fmt.Errorf("unable to create overlay store: %w", err)
	}
	server.OverlayStore = overlayStore

	// routes routes routes
	server.addV1Routes(
		&v1.API{Simulator: server.Simulator},
		&v1.OverlayAPI{Store: server.OverlayStore},
	)

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
