package server

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nsiow/yams/internal/smartrw"
)

func TestSource_Universe_Json(t *testing.T) {
	// Get absolute path to testdata
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	testdataPath := filepath.Join(wd, "..", "..", "testdata", "config-loading", "account_valid.json")

	reader, err := smartrw.NewReader(testdataPath)
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}

	src := &Source{
		Reader:  *reader,
		Refresh: 0,
	}

	universe, err := src.Universe()
	if err != nil {
		t.Fatalf("Universe() error = %v", err)
	}

	if universe == nil {
		t.Fatal("Universe() returned nil")
	}

	// Should have loaded at least one account
	if universe.NumAccounts() == 0 {
		t.Error("Universe() loaded no accounts")
	}
}

func TestSource_Universe_Jsonl(t *testing.T) {
	// Get absolute path to testdata
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	testdataPath := filepath.Join(wd, "..", "..", "testdata", "config-loading", "account_valid.jsonl")

	reader, err := smartrw.NewReader(testdataPath)
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}

	src := &Source{
		Reader:  *reader,
		Refresh: 0,
	}

	universe, err := src.Universe()
	if err != nil {
		t.Fatalf("Universe() error = %v", err)
	}

	if universe == nil {
		t.Fatal("Universe() returned nil")
	}
}

func TestSource_Universe_InvalidFormat(t *testing.T) {
	// Create a temp file with unknown extension
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	reader, err := smartrw.NewReader(tempFile)
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}

	src := &Source{
		Reader:  *reader,
		Refresh: 0,
	}

	_, err = src.Universe()
	if err == nil {
		t.Error("Universe() with invalid format should return error")
	}
}

func TestServer_AddSource(t *testing.T) {
	server, err := NewServer(":8080")
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Get absolute path to testdata
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	testdataPath := filepath.Join(wd, "..", "..", "testdata", "config-loading", "account_valid.json")

	reader, err := smartrw.NewReader(testdataPath)
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}

	src := &Source{
		Reader:  *reader,
		Refresh: 0,
	}

	err = server.AddSource(src)
	if err != nil {
		t.Errorf("AddSource() error = %v", err)
	}

	if len(server.Sources) != 1 {
		t.Errorf("AddSource() sources count = %d, want 1", len(server.Sources))
	}
}

func TestServer_Load(t *testing.T) {
	server, err := NewServer(":8080")
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Get absolute path to testdata
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	testdataPath := filepath.Join(wd, "..", "..", "testdata", "config-loading", "account_valid.json")

	reader, err := smartrw.NewReader(testdataPath)
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}

	src := &Source{
		Reader:  *reader,
		Refresh: 0,
	}

	err = server.Load(src)
	if err != nil {
		t.Errorf("Load() error = %v", err)
	}

	// Verify updated time was set
	if src.Updated.IsZero() {
		t.Error("Load() did not set Updated time")
	}
}

func TestServer_AddSource_WithRefresh(t *testing.T) {
	server, err := NewServer(":8080")
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Get absolute path to testdata
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	testdataPath := filepath.Join(wd, "..", "..", "testdata", "config-loading", "account_valid.json")

	reader, err := smartrw.NewReader(testdataPath)
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}

	src := &Source{
		Reader:  *reader,
		Refresh: 1 * time.Second, // Will start refresh goroutine
	}

	err = server.AddSource(src)
	if err != nil {
		t.Errorf("AddSource() with refresh error = %v", err)
	}

	// Just verify the source was added, refresh goroutine will run in background
	if len(server.Sources) != 1 {
		t.Errorf("AddSource() sources count = %d, want 1", len(server.Sources))
	}
}

func TestServer_Load_InvalidSource(t *testing.T) {
	server, err := NewServer(":8080")
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Create a temp file with invalid content
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.json")
	if err := os.WriteFile(tempFile, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	reader, err := smartrw.NewReader(tempFile)
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}

	src := &Source{
		Reader:  *reader,
		Refresh: 0,
	}

	err = server.Load(src)
	if err == nil {
		t.Error("Load() with invalid source should return error")
	}
}
