package httputil

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func Error(w http.ResponseWriter, req *http.Request, statusCode int, err error) {
	w.WriteHeader(statusCode)

	wrapper := map[string]string{
		"error": err.Error(),
	}

	WriteJsonResponse(w, req, wrapper)
}

func ClientError(w http.ResponseWriter, req *http.Request, err error) {
	Error(w, req, http.StatusBadRequest, err)
}

func ServerError(w http.ResponseWriter, req *http.Request, err error) {
	Error(w, req, http.StatusInternalServerError, err)
}

// TODO(nsiow) implement per-request logging
func WriteJsonResponse(w http.ResponseWriter, req *http.Request, obj any) {
	jsonBytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		slog.Error("error json-ifying object",
			"error", err,
			"obj", obj)
		// Write a simple error response directly to avoid recursion
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}` + "\n"))
		return
	}

	_, err = w.Write(append(jsonBytes, '\n'))
	if err != nil {
		// Don't try to send an error response - the writer is broken.
		// Just log and return.
		slog.Error("error writing response",
			"error", err)
	}
}
