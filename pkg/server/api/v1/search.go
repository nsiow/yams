package v1

import (
	"iter"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/server/httputil"
)

func Search[T entities.Entity](w http.ResponseWriter, req *http.Request, f func() iter.Seq[T]) {
	search := strings.ToLower(req.PathValue("search"))

	keys := []string{}
	for entity := range f() {
		key := entity.Key()
		if strings.Contains(strings.ToLower(key), search) {
			keys = append(keys, key)
		}
	}
	slog.Debug("serving up entities",
		"count", len(keys))

	// sort and return json
	slices.Sort(keys)
	httputil.WriteJsonResponse(w, req, keys)
}
