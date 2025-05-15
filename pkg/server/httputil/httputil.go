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
		ServerError(w, req, err)
		return
	}

	_, err = w.Write(append(jsonBytes, '\n'))
	if err != nil {
		slog.Error("error writing object",
			"error", err)
		ServerError(w, req, err)
		return
	}
}
