package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewServer(t *testing.T) {
	server, err := NewServer(":8080")
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server == nil {
		t.Fatal("NewServer() returned nil server")
	}

	if server.Server == nil {
		t.Fatal("NewServer() server.Server is nil")
	}

	if server.mux == nil {
		t.Fatal("NewServer() server.mux is nil")
	}

	if server.Simulator == nil {
		t.Fatal("NewServer() server.Simulator is nil")
	}

	if server.Addr != ":8080" {
		t.Errorf("NewServer() addr = %s, want :8080", server.Addr)
	}
}

func TestHealthcheck(t *testing.T) {
	server, err := NewServer(":8080")
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/healthcheck", nil)

	server.Healthcheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Healthcheck() status = %d, want %d", w.Code, http.StatusOK)
	}

	want := "{\n  \"status\": \"ok\"\n}\n"
	if w.Body.String() != want {
		t.Errorf("Healthcheck() body = %q, want %q", w.Body.String(), want)
	}
}

func TestStatus(t *testing.T) {
	server, err := NewServer(":8080")
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/status", nil)

	server.Status(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status() status = %d, want %d", w.Code, http.StatusOK)
	}

	var status map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		t.Fatalf("Status() produced invalid JSON: %v", err)
	}

	// Verify required fields exist
	requiredFields := []string{"entities", "accounts", "principals", "groups", "policies", "resources", "sources"}
	for _, field := range requiredFields {
		if _, ok := status[field]; !ok {
			t.Errorf("Status() missing field %q", field)
		}
	}
}

func TestServer_IntegrationRoutes(t *testing.T) {
	server, err := NewServer(":8080")
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Test that routes are registered by making requests through the mux
	tests := []struct {
		method string
		path   string
		want   int
	}{
		{"GET", "/api/v1/healthcheck", http.StatusOK},
		{"GET", "/api/v1/status", http.StatusOK},
		{"GET", "/api/v1/accounts", http.StatusOK},
		{"GET", "/api/v1/groups", http.StatusOK},
		{"GET", "/api/v1/policies", http.StatusOK},
		{"GET", "/api/v1/principals", http.StatusOK},
		{"GET", "/api/v1/resources", http.StatusOK},
		{"GET", "/api/v1/actions", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)

			server.mux.ServeHTTP(w, req)

			if w.Code != tt.want {
				t.Errorf("Request %s %s status = %d, want %d", tt.method, tt.path, w.Code, tt.want)
			}
		})
	}
}
