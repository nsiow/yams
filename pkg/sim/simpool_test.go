package sim

import (
	"context"
	"testing"
	"time"
)

func TestPool_NumWorkers_Default(t *testing.T) {
	ctx := context.Background()
	sim, err := NewSimulator()
	if err != nil {
		t.Fatalf("error creating simulator: %v", err)
	}

	pool := NewPool(ctx, sim)

	// NumWorkers should return a reasonable default
	numWorkers := pool.NumWorkers()
	if numWorkers < 1 {
		t.Fatalf("expected at least 1 worker, got %d", numWorkers)
	}
}

func TestPool_BatchSize_Default(t *testing.T) {
	ctx := context.Background()
	sim, err := NewSimulator()
	if err != nil {
		t.Fatalf("error creating simulator: %v", err)
	}

	pool := NewPool(ctx, sim)

	// BatchSize should return a reasonable default
	batchSize := pool.BatchSize()
	if batchSize < 1 {
		t.Fatalf("expected at least 1 batch size, got %d", batchSize)
	}
}

func TestPool_Timeout_Default(t *testing.T) {
	ctx := context.Background()
	sim, err := NewSimulator()
	if err != nil {
		t.Fatalf("error creating simulator: %v", err)
	}

	pool := NewPool(ctx, sim)

	// Timeout should return a reasonable default
	timeout := pool.Timeout()
	if timeout < 1*time.Second {
		t.Fatalf("expected at least 1 second timeout, got %v", timeout)
	}
}

func TestPool_Start(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sim, err := NewSimulator()
	if err != nil {
		t.Fatalf("error creating simulator: %v", err)
	}

	pool := NewPool(ctx, sim)

	// Start should be idempotent
	pool.Start()
	pool.Start()

	// Just verify the pool can be started
	time.Sleep(50 * time.Millisecond)
}

func TestPool_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	sim, err := NewSimulator()
	if err != nil {
		t.Fatalf("error creating simulator: %v", err)
	}

	pool := NewPool(ctx, sim)
	pool.Start()

	// Cancel context - workers should exit
	cancel()

	// Give workers time to exit
	time.Sleep(100 * time.Millisecond)
}

func TestPool_NumWorkers_FromEnv(t *testing.T) {
	t.Setenv("YAMS_SIM_NUM_WORKERS", "8")

	ctx := context.Background()
	sim, err := NewSimulator()
	if err != nil {
		t.Fatalf("error creating simulator: %v", err)
	}

	pool := NewPool(ctx, sim)
	numWorkers := pool.NumWorkers()

	// Should use env var value
	if numWorkers != 8 {
		t.Errorf("NumWorkers returned %d, want 8", numWorkers)
	}
}

func TestPool_BatchSize_FromEnv(t *testing.T) {
	t.Setenv("YAMS_SIM_BATCH_SIZE", "512")

	ctx := context.Background()
	sim, err := NewSimulator()
	if err != nil {
		t.Fatalf("error creating simulator: %v", err)
	}

	pool := NewPool(ctx, sim)
	batchSize := pool.BatchSize()

	if batchSize != 512 {
		t.Errorf("BatchSize returned %d, want 512", batchSize)
	}
}

func TestPool_Timeout_FromEnv(t *testing.T) {
	t.Setenv("YAMS_SIM_TIMEOUT", "120")

	ctx := context.Background()
	sim, err := NewSimulator()
	if err != nil {
		t.Fatalf("error creating simulator: %v", err)
	}

	pool := NewPool(ctx, sim)
	timeout := pool.Timeout()

	if timeout != 120*time.Second {
		t.Errorf("Timeout returned %v, want 120s", timeout)
	}
}
