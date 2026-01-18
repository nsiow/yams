package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/nsiow/yams/cmd/yams/cli"
	"github.com/nsiow/yams/internal/smartrw"
)

func TestNewServer(t *testing.T) {
	server, err := NewServer(&cli.Flags{Addr: ":8080"})
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
	server, err := NewServer(&cli.Flags{Addr: ":8080"})
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
	server, err := NewServer(&cli.Flags{Addr: ":8080"})
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

func TestStatus_WithSources(t *testing.T) {
	server, err := NewServer(&cli.Flags{Addr: ":8080"})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Add a mock source directly to test status with sources
	server.Sources = append(server.Sources, &Source{
		Reader:  smartrw.Reader{Source: "test-source.json"},
		Refresh: 0,
		Updated: time.Now(),
	})

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

	// Verify sources field contains our source
	sources, ok := status["sources"].([]any)
	if !ok || len(sources) != 1 {
		t.Errorf("Status() sources field incorrect, got %v", status["sources"])
	}
}

func TestServer_IntegrationRoutes(t *testing.T) {
	server, err := NewServer(&cli.Flags{Addr: ":8080"})
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

func TestCorsMiddleware(t *testing.T) {
	// Create a simple handler that the middleware wraps
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Wrap with cors middleware
	wrapped := corsMiddleware(handler)

	t.Run("normal_request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)

		wrapped.ServeHTTP(w, req)

		// Check CORS headers are set
		if w.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Error("expected Access-Control-Allow-Origin: *")
		}
		if w.Header().Get("Access-Control-Allow-Methods") != "*" {
			t.Error("expected Access-Control-Allow-Methods: *")
		}
		if w.Header().Get("Access-Control-Allow-Headers") != "*" {
			t.Error("expected Access-Control-Allow-Headers: *")
		}
		if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
			t.Error("expected Access-Control-Allow-Credentials: true")
		}

		// Should pass through to handler
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("options_preflight", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("OPTIONS", "/test", nil)

		wrapped.ServeHTTP(w, req)

		// Check CORS headers are set
		if w.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Error("expected Access-Control-Allow-Origin: *")
		}

		// OPTIONS should return 200 without calling the underlying handler
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Body should be empty for OPTIONS
		if w.Body.Len() > 0 {
			t.Error("OPTIONS response body should be empty")
		}
	})
}

func TestStatus_WithEnvVars(t *testing.T) {
	// Set an env var to test the env section
	envKey := "YAMS_TEST_VAR"
	os.Setenv(envKey, "test_value")
	defer os.Unsetenv(envKey)

	server, err := NewServer(&cli.Flags{Addr: ":8080", Env: []string{envKey}})
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

	// Verify env field is present
	env, ok := status["env"].(map[string]any)
	if !ok {
		t.Fatal("Status() missing env field")
	}

	if env[envKey] != "test_value" {
		t.Errorf("Status() env[%s] = %v, want 'test_value'", envKey, env[envKey])
	}
}

func TestStatus_WithEmptyEnvVar(t *testing.T) {
	// Set an empty env var - should not be included
	envKey := "YAMS_TEST_EMPTY_VAR"
	os.Setenv(envKey, "")
	defer os.Unsetenv(envKey)

	server, err := NewServer(&cli.Flags{Addr: ":8080", Env: []string{envKey}})
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

	// Verify env field is not present (empty env vars are excluded)
	if _, ok := status["env"]; ok {
		t.Error("Status() should not include env field when all env vars are empty")
	}
}
