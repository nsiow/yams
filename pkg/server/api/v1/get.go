package v1

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/server/httputil"
)

func Get[T entities.Entity](w http.ResponseWriter, req *http.Request, f func(string) (T, bool)) {
	// parse path variables
	key := req.PathValue("key")
	if len(key) == 0 {
		httputil.ClientError(w, req, fmt.Errorf("no ARN provided"))
		return
	}

	// handle freeze suffix
	var freeze bool
	if strings.HasSuffix(key, "/freeze") {
		key = strings.TrimSuffix(key, "/freeze")
		freeze = true
	}

	// lookup entity
	entity, ok := f(key)
	if !ok {
		httputil.Error(w, req, http.StatusNotFound, fmt.Errorf("unable to find entity: '%s'", key))
		return
	}

	// handle freeze if needed
	var obj any
	var err error
	if freeze {
		obj, err = entity.Repr()
		if err != nil {
			slog.Error("error freezing entity", "error", err)
			httputil.ServerError(w, req, err)
			return
		}
	} else {
		obj = entity
	}

	httputil.WriteJsonResponse(w, req, obj)
}
