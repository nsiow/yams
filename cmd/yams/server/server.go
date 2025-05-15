package server

import (
	"log/slog"
	"time"

	"github.com/nsiow/yams/cmd/yams/cli"
	"github.com/nsiow/yams/internal/smartrw"
	"github.com/nsiow/yams/pkg/server"
)

// Logic for the "server" subcommand
func Run(opts *cli.Flags) {
	srv, err := server.NewServer(opts.Addr)
	if err != nil {
		cli.Fail("error starting server: %v", err)
	}

	for _, src := range opts.Sources {
		reader, err := smartrw.NewReader(src)
		if err != nil {
			cli.Fail("error when initializing reader for source '%s': %v", src, err)
		}

		source := server.Source{
			Reader:  *reader,
			Refresh: time.Second * time.Duration(opts.Refresh),
		}

		err = srv.AddSource(&source)
		if err != nil {
			cli.Fail("error attempting to add source '%s': %v", src, err)
		}
	}

	slog.Info("starting server")
	err = srv.ListenAndServe()
	if err != nil {
		cli.Fail("error from server: %v", err)
	}
}
