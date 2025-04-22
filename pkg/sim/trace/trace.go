package trace

// Trace is a stack-based logger which allows us to follow a simulation execution
type Trace struct {
	enabled bool

	stack []*frame
}

// curr returns a pointer to the current (topmost) frame
func (t *Trace) curr() *frame {
	if len(t.stack) <= 0 {
		panic("attempt to look up current frame for empty stack")
	}

	return t.stack[len(t.stack)-1]
}

// New creates and returns a new Trace with a root frame initialized
func New() *Trace {
	root := frame{header: "root"}
	t := Trace{
		stack: []*frame{
			&root,
		},
	}
	return &t
}

// Enable turns on recording for this Trace
func (t *Trace) Enable() {
	t.enabled = true
}

// Disable turns off recording for this Trace
func (t *Trace) Disable() {
	t.enabled = false
}

// Push creates a new frame and adds it to the top of the stack
func (t *Trace) Push(header string, args ...any) {
	if !t.enabled {
		return
	}

	subframe := t.curr().subframe(header, args...)
	t.stack = append(t.stack, subframe)
}

// Pop removes the topmost frame from the trace and saves it
func (t *Trace) Pop() {
	if !t.enabled {
		return
	}

	if len(t.stack) <= 1 {
		panic("attempt to pop root frame from trace stack")
	}

	t.stack = t.stack[:len(t.stack)-1]
}

// Observation records a single record about comparison (e.g. "I compared these two things")
func (t *Trace) Observation(msg string, args ...any) {
	if !t.enabled {
		return
	}

	t.curr().emit(msg, args...)
}

// Decision records a single record about an access decision (e.g. "This resulted in Effect=Allow")
func (t *Trace) Decision(msg string, args ...any) {
	if !t.enabled {
		return
	}

	t.curr().emit(msg, args...)
}
