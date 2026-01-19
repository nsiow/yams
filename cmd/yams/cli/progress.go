package cli

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var spinnerFrames = []string{"|", "/", "-", "\\"}

// Progress provides a simple progress indicator with spinner
type Progress struct {
	message string
	stop    chan struct{}
	done    chan struct{}
	mu      sync.Mutex
	active  bool
}

// NewProgress creates a new progress indicator
func NewProgress(message string) *Progress {
	return &Progress{
		message: message,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
}

// Start begins the progress indicator if stderr is a TTY
func (p *Progress) Start() {
	if !StderrIsTTY() {
		return
	}

	p.mu.Lock()
	if p.active {
		p.mu.Unlock()
		return
	}
	p.active = true
	p.mu.Unlock()

	go func() {
		defer close(p.done)
		i := 0
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-p.stop:
				fmt.Fprintf(os.Stderr, "\r\033[K") // clear line
				return
			case <-ticker.C:
				frame := spinnerFrames[i%len(spinnerFrames)]
				fmt.Fprintf(os.Stderr, "\r%s %s", frame, p.message)
				i++
			}
		}
	}()
}

// Update changes the progress message
func (p *Progress) Update(message string) {
	p.mu.Lock()
	p.message = message
	p.mu.Unlock()
}

// Stop halts the progress indicator
func (p *Progress) Stop() {
	p.mu.Lock()
	if !p.active {
		p.mu.Unlock()
		return
	}
	p.active = false
	p.mu.Unlock()

	close(p.stop)
	<-p.done
}

// StopWithMessage stops the spinner and prints a final message
func (p *Progress) StopWithMessage(message string) {
	p.Stop()
	if StderrIsTTY() {
		fmt.Fprintln(os.Stderr, message)
	}
}
