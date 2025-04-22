package trace

// Trace is a stack-based logger which allows us to follow a simulation execution
type Trace struct {
	enabled bool

	depth int
	stack []*Frame
	hist  []*Frame
}

// curr returns the current operating frame for this trace
func (t *Trace) curr() *Frame {
	if len(t.stack) == 0 {
		panic("somehow reached empty stack")
	}

	return t.stack[len(t.stack)-1]
}

// New creates and returns a new Trace with a root frame initialized
func New() *Trace {
	t := Trace{}
	t.Push("root")
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
func (t *Trace) Push(frame string, args ...any) {
	if !t.enabled {
		return
	}

	t.stack = append(t.stack, NewFrame(t.depth, frame, args...))
	t.depth += 1
}

// Pop removes the topmost frame from the trace and saves it
func (t *Trace) Pop() {
	if !t.enabled {
		return
	}

	if t.depth -= 1; t.depth < 0 {
		panic("stack underflow for trace")
	}

	t.hist = append(t.hist, t.stack[len(t.stack)-1])
	t.stack = t.stack[:len(t.stack)-1]
}

// Observation records a single record about comparison (e.g. "I compared these two things")
func (t *Trace) Observation(msg string, args ...any) {
	if !t.enabled {
		return
	}

	t.curr().record(format(msg, args...))
}

// Decision records a single record about an access decision (e.g. "This resulted in Effect=Allow")
func (t *Trace) Decision(msg string, args ...any) {
	if !t.enabled {
		return
	}

	t.curr().record(format(msg, args...))
}

// History returns all frames saved by the trace, in sequential order
func (t *Trace) History() []*Frame {
	return t.stack
}
