package server

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nsiow/yams/cmd/yams/cli"
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
	server, err := NewServer(&cli.Flags{Addr: ":8080"})
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
	server, err := NewServer(&cli.Flags{Addr: ":8080"})
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
	server, err := NewServer(&cli.Flags{Addr: ":8080"})
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
	server, err := NewServer(&cli.Flags{Addr: ":8080"})
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

func TestServer_AddSource_Invalid(t *testing.T) {
	server, err := NewServer(&cli.Flags{Addr: ":8080"})
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

	// AddSource should return error because Load fails
	err = server.AddSource(src)
	if err == nil {
		t.Error("AddSource() with invalid source should return error")
	}

	// Sources list should be empty
	if len(server.Sources) != 0 {
		t.Errorf("AddSource() with error should not add source, got %d sources", len(server.Sources))
	}
}

func TestServer_Refresh_Success(t *testing.T) {
	server, err := NewServer(&cli.Flags{Addr: ":8080"})
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
		Refresh: 50 * time.Millisecond, // Very short refresh for testing
	}

	// Load the source first (so we can test refresh)
	err = server.Load(src)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	initialUpdate := src.Updated

	// Start refresh in goroutine
	go server.Refresh(src)

	// Wait for at least one refresh cycle
	time.Sleep(100 * time.Millisecond)

	// Check that refresh updated the timestamp
	if !src.Updated.After(initialUpdate) {
		t.Error("Refresh() did not update source timestamp")
	}
}

func TestServer_Refresh_ResetError(t *testing.T) {
	server, err := NewServer(&cli.Flags{Addr: ":8080"})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Create a temp file that we'll delete to cause reset error
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.json")
	validJSON := `[{"resourceType": "Yams::Organizations::Account", "accountId": "123456789012", "arn": "arn:aws:::account/123456789012"}]`
	if err := os.WriteFile(tempFile, []byte(validJSON), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	reader, err := smartrw.NewReader(tempFile)
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}

	src := &Source{
		Reader:  *reader,
		Refresh: 50 * time.Millisecond,
	}

	// Load the source first
	err = server.Load(src)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Delete the file to cause Reset() to fail
	os.Remove(tempFile)

	// Start refresh - should exit after first failed reset
	go server.Refresh(src)

	// Wait for the refresh to attempt and fail
	time.Sleep(100 * time.Millisecond)

	// The goroutine should have exited due to reset error
	// (no way to verify directly, but coverage will show the path was taken)
}

func TestServer_Refresh_LoadError(t *testing.T) {
	server, err := NewServer(&cli.Flags{Addr: ":8080"})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Create a temp file that we'll corrupt after first load
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.json")
	validJSON := `[{"resourceType": "Yams::Organizations::Account", "accountId": "123456789012", "arn": "arn:aws:::account/123456789012"}]`
	if err := os.WriteFile(tempFile, []byte(validJSON), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	reader, err := smartrw.NewReader(tempFile)
	if err != nil {
		t.Fatalf("failed to create reader: %v", err)
	}

	src := &Source{
		Reader:  *reader,
		Refresh: 50 * time.Millisecond,
	}

	// Load the source first
	err = server.Load(src)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Corrupt the file to cause Load() to fail on refresh
	os.WriteFile(tempFile, []byte("invalid json"), 0644)

	// Start refresh - should exit after first failed load
	go server.Refresh(src)

	// Wait for the refresh to attempt and fail
	time.Sleep(100 * time.Millisecond)

	// The goroutine should have exited due to load error
}

func TestSource_Universe_LoadJsonlError(t *testing.T) {
	// Create a temp file with invalid jsonl content
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.jsonl")
	if err := os.WriteFile(tempFile, []byte("invalid jsonl content"), 0644); err != nil {
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
		t.Error("Universe() with invalid jsonl should return error")
	}
}
