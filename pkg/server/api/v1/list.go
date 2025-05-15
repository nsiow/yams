package v1

import (
	"iter"
	"log/slog"
	"net/http"
	"slices"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/server/httputil"
)

func List[T entities.Entity](w http.ResponseWriter, req *http.Request, f func() iter.Seq[T]) {
	keys := []string{}
	for entity := range f() {
		keys = append(keys, entity.Key())
	}
	slog.Debug("serving up entities",
		"count", len(keys))

	// sort and return json
	slices.Sort(keys)
	httputil.WriteJsonResponse(w, req, keys)
}
