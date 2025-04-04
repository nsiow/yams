package trace

import (
	"fmt"
	"maps"
	"strings"
)

// Trace is a stack-based, leveled, rich logger which allows us to follow a simulation execution
type Trace struct {
	minLevel int
	stack    []string
	attrs    []map[string]any
	buf      []Record
}

// New creates and returns a new Trace with a root frame initialized
func New() *Trace {
	trc := &Trace{} // TODO(nsiow) maybe preallocate?
	trc.Push("root")
	return trc
}

// Push creates a new frame and attribute set at current depth + 1
func (t *Trace) Push(frame string) {
	t.stack = append(t.stack, frame)
	t.attrs = append(t.attrs, make(map[string]any))
}

// Pop removes the topmost (higest depth) frame and attribute set
func (t *Trace) Pop() {
	if len(t.stack) > 1 {
		t.stack = t.stack[:len(t.stack)-1]
	}
	if len(t.attrs) > 1 {
		t.attrs = t.attrs[:len(t.attrs)-1]
	}
}

// Attr creates a new attribute at the topmost frame
//
// It will NOT be inherited by tracing frames at a higher or lower depth
func (t *Trace) Attr(k string, v any) {
	t.attrs[len(t.attrs)-1][k] = v
}

// History returns all records saved by the trace, in sequential order
func (t *Trace) History() []Record {
	return t.buf
}

// Log returns the history of the trace in a string-based, human-readable format
func (t *Trace) Log() string {
	log := []string{}

	for _, record := range t.History() {
		s := fmt.Sprintf("depth=(%d) frame=(%s) message=(%s) attrs=(%+v)",
			record.Depth, record.Frame, record.Message, record.Attrs)
		log = append(log, s)
	}

	return strings.Join(log, "\n")
}

// SetLevel assigns the minimum "logging" level of the trace object
//
// For example, setting LEVEL_COMPARISON will mean that comparison AND decision records are kept
// Setting LEVEL_DECISION will mean that comparison records are no longer kept
func (t *Trace) SetLevel(l Level) {
	t.minLevel = l
}

// Observation records a single record about comparison (e.g. "I compared these two things")
func (t *Trace) Observation(msg string) {
	if t.minLevel <= LEVEL_OBSERVATION {
		t.save(msg)
	}
}

// Decision records a single record about an access decision (e.g. "This resulted in Effect=Allow")
func (t *Trace) Decision(msg string) {
	if t.minLevel <= LEVEL_DECISION {
		t.save(msg)
	}
}

// copyAttr is a helper function which makes a copy of the provided attributes
func (t *Trace) copyAttr(m map[string]any) map[string]any {
	c := make(map[string]any)
	maps.Copy(c, m)
	return c
}

// save creates a new trace.Record object using the topmost frame, attributes, and provided message
func (t *Trace) save(msg string) {
	r := Record{
		Message: msg,
		Attrs:   t.copyAttr(t.attrs[len(t.attrs)-1]),
		Frame:   t.stack[len(t.stack)-1],
		Depth:   len(t.stack) - 1,
	}
	t.buf = append(t.buf, r)
}
