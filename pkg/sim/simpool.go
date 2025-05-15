package sim

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// -------------------------------------------------------------------------------------------------
// Types
// -------------------------------------------------------------------------------------------------

type simIn struct {
	AuthContext AuthContext
	Options     Options
}

type simOut struct {
	Result *SimResult
	Error  error
}

type simBatch struct {
	Jobs     []simIn
	Finished chan<- simOut
}

type Pool struct {
	Simulator *Simulator // TODO(nsiow) revisit this: make default Pool useful?
	Ctx       context.Context

	numWorkers int
	batchSize  int
	timeout    time.Duration

	started sync.Once
	work    chan simBatch
}

// -------------------------------------------------------------------------------------------------
// Pool Configuration
// -------------------------------------------------------------------------------------------------

func NewPool(ctx context.Context, simulator *Simulator) *Pool {
	return &Pool{
		Simulator: simulator,
		Ctx:       ctx,
		work:      make(chan simBatch, 512), // TODO(nsiow) figure out what a good default is
	}
}

func (p *Pool) NumWorkers() int {
	if p.numWorkers == 0 {
		fromEnv := os.Getenv("YAMS_SIM_NUM_WORKERS")
		num, err := strconv.Atoi(fromEnv)
		if err == nil {
			p.numWorkers = num
		}

		p.numWorkers = runtime.NumCPU() // default to some reasonable number of workers
	}

	return p.numWorkers
}

func (p *Pool) SetWorkers(num int) {
	p.numWorkers = num
}

func (p *Pool) BatchSize() int {
	if p.batchSize == 0 {
		fromEnv := os.Getenv("YAMS_SIM_BATCH_SIZE")
		num, err := strconv.Atoi(fromEnv)
		if err == nil {
			p.batchSize = num
		}

		p.batchSize = 1024 // default to some reasonable batch size
	}

	return p.batchSize
}

func (p *Pool) SetBatchSize(num int) {
	p.batchSize = num
}

func (p *Pool) Timeout() time.Duration {
	if p.timeout == 0 {
		fromEnv := os.Getenv("YAMS_SIM_TIMEOUT")
		num, err := strconv.Atoi(fromEnv)
		if err == nil {
			p.timeout = time.Duration(num * int(time.Second))
		}

		p.timeout = 30 * time.Second
	}

	return p.timeout
}

func (p *Pool) SetTimeout(timeout time.Duration) {
	p.timeout = timeout
}

// -------------------------------------------------------------------------------------------------
// Pool Execution
// -------------------------------------------------------------------------------------------------

func (p *Pool) Start() {
	p.started.Do(func() {
		if p.numWorkers < 1 {
			p.numWorkers = 1
		}

		for range p.numWorkers {
			go p.startWorker()
		}
	})
}

func (p *Pool) startWorker() {
	for {
		select {
		case <-p.Ctx.Done():
			return
		case batch := <-p.work:
			p.handleBatch(batch)
		}
	}
}

func (p *Pool) handleBatch(b simBatch) {
	for _, item := range b.Jobs {
		result, err := p.Simulator.SimulateWithOptions(item.AuthContext, item.Options)
		out := simOut{Result: result, Error: err}
		b.Finished <- out
	}
}

func (p *Pool) Submit(b simBatch) {
	p.Start()
	p.work <- b
}
