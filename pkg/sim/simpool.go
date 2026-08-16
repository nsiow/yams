package sim

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"time"
)

const (
	defaultBatchSize      = 1024
	defaultTimeoutSeconds = 60
)

type Pool struct {
	Simulator *Simulator // TODO(nsiow) revisit this: make default Pool useful?
	Ctx       context.Context

	numWorkers int
	batchSize  int
	timeout    time.Duration
}

// -------------------------------------------------------------------------------------------------
// Pool Configuration
// -------------------------------------------------------------------------------------------------

func NewPool(ctx context.Context, simulator *Simulator) *Pool {
	p := &Pool{
		Simulator:  simulator,
		Ctx:        ctx,
		numWorkers: positiveEnvInt("YAMS_SIM_NUM_WORKERS", runtime.NumCPU()),
		batchSize:  positiveEnvInt("YAMS_SIM_BATCH_SIZE", defaultBatchSize),
		timeout: time.Duration(
			positiveEnvInt("YAMS_SIM_TIMEOUT", defaultTimeoutSeconds),
		) * time.Second,
	}

	slog.Info("created pool",
		"num_workers", p.NumWorkers(),
		"batch_size", p.BatchSize(),
		"timeout", p.Timeout())
	return p
}

func (p *Pool) NumWorkers() int {
	return p.numWorkers
}

// BatchSize is the target number of evaluations claimed at a time by the streaming scheduler.
func (p *Pool) BatchSize() int {
	return p.batchSize
}

func (p *Pool) Timeout() time.Duration {
	return p.timeout
}

func positiveEnvInt(name string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

// -------------------------------------------------------------------------------------------------
// Pool Execution
// -------------------------------------------------------------------------------------------------

// Start is retained for compatibility. Workers are now scoped to individual product calls.
// Deprecated: callers no longer need to start the pool.
func (*Pool) Start() {}
