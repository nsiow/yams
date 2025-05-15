package server

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/nsiow/yams/internal/smartrw"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/loaders/awsconfig"
)

type Source struct {
	Reader  smartrw.Reader
	Refresh time.Duration
	Updated time.Time
}

func (s *Source) Universe() (*entities.Universe, error) {
	loader := awsconfig.NewLoader()
	var err error

	// TODO(nsiow) implement this more accurately, allow non-Config sources
	if strings.Contains(s.Reader.Source, ".jsonl") {
		err = loader.LoadJsonl(s.Reader)
	} else if strings.Contains(s.Reader.Source, ".json") {
		err = loader.LoadJson(s.Reader)
	} else {
		return nil, fmt.Errorf("unsure what loader to use for source: %s", s.Reader.Source)
	}

	if err != nil {
		return nil, err
	}

	err = s.Reader.Close()
	if err != nil {
		return nil, err
	}

	return loader.Universe(), nil
}

func (serv *Server) AddSource(src *Source) error {
	slog.Info("initial loading of source",
		"source", src.Reader.Source)

	err := serv.Load(src)
	if err != nil {
		return err
	}

	serv.Sources = append(serv.Sources, src)

	if src.Refresh > 0 {
		slog.Info("scheduling source to refresh",
			"source", src.Reader.Source,
			"refreshEvery", src.Refresh)
		go serv.Refresh(src)
	}

	return nil
}

func (serv *Server) Load(src *Source) error {
	slog.Info("loading source",
		"source", src.Reader.Source)

	uv, err := src.Universe()
	if err != nil {
		slog.Error("error loading source",
			"source", src,
			"universe", uv)
		return err
	}

	slog.Info("finished loading items",
		"numLoaded", uv.Size())

	serv.Simulator.Universe.Merge(uv)
	src.Updated = time.Now()

	slog.Info("universe after loading",
		"size", serv.Simulator.Universe.Size())

	return nil
}

func (serv *Server) Refresh(src *Source) {
	for tick := range time.Tick(src.Refresh) {
		slog.Info("refreshing source",
			"tick", tick,
			"source", src.Reader.Source)

		err := src.Reader.Reset()
		if err != nil {
			slog.Error("error resetting source",
				"tick", tick,
				"source", src.Reader.Source,
				"error", err)
			return
		}

		err = serv.Load(src)
		if err != nil {
			slog.Error("error refreshing source",
				"tick", tick,
				"source", src.Reader.Source,
				"error", err)
			return
		}
	}
}
