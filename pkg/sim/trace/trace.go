package trace

import (
	"fmt"
	"strings"
)

// -------------------------------------------------------------------------------------------------
// Traces
// -------------------------------------------------------------------------------------------------

// Trace is a stack-based logger which allows us to follow a simulation execution
type Trace struct {
	enabled bool

	stack []*Frame
}

// curr returns a pointer to the current (topmost) Frame
func (t *Trace) curr() *Frame {
	if len(t.stack) <= 0 {
		panic("attempt to look up current frame for empty stack")
	}

	return t.stack[len(t.stack)-1]
}

// New creates and returns a new Trace with a root Frame initialized
func New() *Trace {
	root := Frame{
		Header: "root",
		Depth:  0,
	}

	t := Trace{
		stack: []*Frame{
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

// Push creates a new Frame and adds it to the top of the stack
func (t *Trace) Push(header string, args ...any) {
	if !t.enabled {
		return
	}

	subFrame := t.curr().subFrame(header, args...)
	t.stack = append(t.stack, subFrame)
}

// Pop removes the topmost Frame from the trace and saves it
func (t *Trace) Pop() {
	if !t.enabled {
		return
	}

	if len(t.stack) <= 1 {
		panic("attempt to pop root frame from trace stack")
	}

	t.stack = t.stack[:len(t.stack)-1]
}

// Log records a single record about comparison (e.g. "I compared these two things")
func (t *Trace) Log(msg string, args ...any) {
	if !t.enabled {
		return
	}

	t.curr().emit(msg, args...)
}

// Allowed marks the current frame as associated with an ALLOW-type decision
func (t *Trace) Allowed(msg string, args ...any) {
	if !t.enabled {
		return
	}

	t.curr().set("allowed", "true")
	t.curr().emit(msg, args...)
}

// Denied marks the current frame as associated with an DENY-type decision
func (t *Trace) Denied(msg string, args ...any) {
	if !t.enabled {
		return
	}

	t.curr().set("denied", "true")
	t.curr().emit(msg, args...)
}

// Walk recursively walks the [Frame] objects emitted by the [Trace] and calls the provided function
// for each emitted event
func (t *Trace) Walk(w Walker) {
	if len(t.stack) == 0 {
		panic("trace somehow has empty stack")
	}

	walk(w, t.stack[0])
}

// walk is the internal helper function for [Trace.Walk]
func walk(w Walker, fr *Frame) {
	w.FrameStart(fr)
	defer w.FrameEnd(fr)

	for _, evt := range fr.hist {
		switch evt.eventType {
		case eventTypeMessage:
			w.Message(fr, evt.message)
		case eventTypeSubFrame:
			walk(w, evt.subFrame)
		default:
			panic(fmt.Sprintf("unexpected event type: %s", evt.eventType))
		}
	}
}

// String returns the [Trace] formatted as a barebones human-readable string
func (t *Trace) String() string {
	p := Printer{}
	t.Walk(&p)
	return p.Print()
}

// -------------------------------------------------------------------------------------------------
// Walker
// -------------------------------------------------------------------------------------------------

// Walker defines the interface for a type that is able to recursively walk a [Trace] execution
type Walker interface {
	FrameStart(*Frame)
	Message(*Frame, string)
	FrameEnd(*Frame)
}

// -------------------------------------------------------------------------------------------------
// Printer
// -------------------------------------------------------------------------------------------------

// Printer is a specialized [Walker] which collects frames and messages as strings as it walks the
// Trace execution and presents them in human-readable form
type Printer struct {
	messages []string
}

func (p *Printer) Add(s string) {
	p.messages = append(p.messages, s)
}

func (p *Printer) Print() string {
	return strings.Join(p.messages, "\n")
}

func (p *Printer) Annotate(fr *Frame) string {
	if fr.Attributes["allowed"] == "true" {
		return "(allow) "
	}

	if fr.Attributes["denied"] == "true" {
		return "(deny) "
	}

	return ""
}

func (p *Printer) Indent(fr *Frame) string {
	return strings.Repeat("  ", fr.Depth)
}

func (p *Printer) FrameStart(fr *Frame) {
	p.Add(
		fmt.Sprintf("%s%sbegin: %s", p.Indent(fr), p.Annotate(fr), fr.Header),
	)
}

func (p *Printer) Message(fr *Frame, msg string) {
	p.Add(
		fmt.Sprintf("%s  %s", p.Indent(fr), msg),
	)
}

func (p *Printer) FrameEnd(fr *Frame) {
	p.Add(
		fmt.Sprintf("%send: %s", p.Indent(fr), fr.Header),
	)
}

// -------------------------------------------------------------------------------------------------
// Frames
// -------------------------------------------------------------------------------------------------

// Frame represents a logical span of evaluation logic, i.e. "evaluating resource policies"
type Frame struct {
	Header     string
	Depth      int
	Attributes map[string]string
	hist       []event
}

func (f *Frame) set(key, value string) {
	if f.Attributes == nil {
		f.Attributes = make(map[string]string)
	}

	f.Attributes[key] = value
}

func (f *Frame) emit(msg string, args ...any) {
	next := event{
		eventType: eventTypeMessage,
		message:   format(msg, args...),
	}
	f.hist = append(f.hist, next)
}

func (f *Frame) subFrame(header string, args ...any) *Frame {
	subFrame := Frame{
		Header: format(header, args...),
		Depth:  f.Depth + 1,
	}
	next := event{
		eventType: eventTypeSubFrame,
		subFrame:  &subFrame,
	}
	f.hist = append(f.hist, next)
	return &subFrame
}

// -------------------------------------------------------------------------------------------------
// Events
// -------------------------------------------------------------------------------------------------

// eventType is an enum representing the different type of events we may emit
type eventType string

const (
	eventTypeMessage  eventType = "EVT_TYPE_MESSAGE"
	eventTypeSubFrame eventType = "EVT_TYPE_SUBFRAME"
)

// event represents an emitted event
type event struct {
	eventType eventType // always present

	message  string // only present when eventType == eventTypeMessage
	subFrame *Frame // only present when eventType == eventTypeSubFrame
}

// -------------------------------------------------------------------------------------------------
// Formatting
// -------------------------------------------------------------------------------------------------

// format is a helper function for sprintf-style formatting of records
//
// This is factored out for two reasons:
// - avoid allocations for arg-less formatting calls
// - provide a single place for any formatting customizations
func format(msg string, args ...any) string {
	if len(args) == 0 {
		return msg
	}

	return fmt.Sprintf(msg, args...)
}
