package httputil

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		err        error
		wantBody   map[string]string
	}{
		{
			name:       "bad request error",
			statusCode: http.StatusBadRequest,
			err:        errors.New("bad request"),
			wantBody:   map[string]string{"error": "bad request"},
		},
		{
			name:       "not found error",
			statusCode: http.StatusNotFound,
			err:        errors.New("not found"),
			wantBody:   map[string]string{"error": "not found"},
		},
		{
			name:       "internal server error",
			statusCode: http.StatusInternalServerError,
			err:        errors.New("internal error"),
			wantBody:   map[string]string{"error": "internal error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)

			Error(w, req, tt.statusCode, tt.err)

			if w.Code != tt.statusCode {
				t.Errorf("Error() status = %d, want %d", w.Code, tt.statusCode)
			}

			var got map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if got["error"] != tt.wantBody["error"] {
				t.Errorf("Error() body = %v, want %v", got, tt.wantBody)
			}
		})
	}
}

func TestClientError(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	testErr := errors.New("client error")

	ClientError(w, req, testErr)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ClientError() status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if got["error"] != "client error" {
		t.Errorf("ClientError() body = %v, want error='client error'", got)
	}
}

func TestServerError(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	testErr := errors.New("server error")

	ServerError(w, req, testErr)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("ServerError() status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if got["error"] != "server error" {
		t.Errorf("ServerError() body = %v, want error='server error'", got)
	}
}

func TestWriteJsonResponse(t *testing.T) {
	tests := []struct {
		name string
		obj  any
	}{
		{
			name: "simple object",
			obj:  map[string]string{"key": "value"},
		},
		{
			name: "array",
			obj:  []string{"a", "b", "c"},
		},
		{
			name: "nested object",
			obj: map[string]any{
				"name":  "test",
				"count": 42,
				"items": []string{"x", "y"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)

			WriteJsonResponse(w, req, tt.obj)

			// Verify response is valid JSON
			var got any
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("WriteJsonResponse() produced invalid JSON: %v", err)
			}

			// Re-marshal both and compare
			wantBytes, _ := json.Marshal(tt.obj)
			gotBytes, _ := json.Marshal(got)

			var wantNorm, gotNorm any
			json.Unmarshal(wantBytes, &wantNorm)
			json.Unmarshal(gotBytes, &gotNorm)

			wantFinal, _ := json.Marshal(wantNorm)
			gotFinal, _ := json.Marshal(gotNorm)

			if string(wantFinal) != string(gotFinal) {
				t.Errorf("WriteJsonResponse() = %s, want %s", gotFinal, wantFinal)
			}
		})
	}
}

// errWriter is a writer that always fails
type errWriter struct {
	header http.Header
}

func (e *errWriter) Header() http.Header {
	if e.header == nil {
		e.header = make(http.Header)
	}
	return e.header
}

func (e *errWriter) Write([]byte) (int, error) {
	return 0, errors.New("write error")
}

func (e *errWriter) WriteHeader(int) {}

func TestWriteJsonResponse_WriteError(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	w := &errWriter{}

	// This should not panic or loop infinitely, just log the error
	WriteJsonResponse(w, req, map[string]string{"key": "value"})
}

func TestWriteJsonResponse_MarshalError(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	// Create an object that can't be marshaled (channel)
	obj := make(chan int)

	WriteJsonResponse(w, req, obj)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("WriteJsonResponse with unmarshalable object status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
