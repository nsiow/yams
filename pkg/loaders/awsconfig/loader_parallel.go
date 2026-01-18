// pkg/loaders/awsconfig/loader_parallel.go
package awsconfig

import (
	"encoding/json"
	"runtime"
	"sync"
)

// loadJob represents a single item to be loaded by a worker
type loadJob struct {
	typ string
	raw json.RawMessage
}

// loadError wraps an error with the raw data that caused it for debugging
type loadError struct {
	err error
	raw json.RawMessage
}

// loaderPool manages parallel loading of config items
type loaderPool struct {
	loader     *Loader
	numWorkers int
	jobs       chan loadJob
	wg         sync.WaitGroup

	// Error handling: capture first error
	errOnce sync.Once
	err     loadError
}

// newLoaderPool creates a new pool for parallel item loading
func newLoaderPool(loader *Loader) *loaderPool {
	numWorkers := runtime.NumCPU()
	return &loaderPool{
		loader:     loader,
		numWorkers: numWorkers,
		jobs:       make(chan loadJob, numWorkers*256),
	}
}

// start launches the worker goroutines
func (p *loaderPool) start() {
	for range p.numWorkers {
		p.wg.Add(1)
		go p.worker()
	}
}

// worker processes jobs from the channel until it's closed
func (p *loaderPool) worker() {
	defer p.wg.Done()

	for job := range p.jobs {
		blob := configBlob{
			Type: job.typ,
			raw:  job.raw,
		}

		err := p.loader.loadItem(blob)
		if err != nil {
			p.errOnce.Do(func() {
				p.err = loadError{err: err, raw: job.raw}
			})
		}
	}
}

// submit adds a job to the queue
func (p *loaderPool) submit(typ string, raw json.RawMessage) {
	p.jobs <- loadJob{typ: typ, raw: raw}
}

// close signals no more jobs and waits for completion
func (p *loaderPool) close() {
	close(p.jobs)
	p.wg.Wait()
}

// error returns any error that occurred during processing
func (p *loaderPool) error() error {
	if p.err.err != nil {
		return p.err.err
	}
	return nil
}
