package trace

// -------------------------------------------------------------------------------------------------
// Frames
// -------------------------------------------------------------------------------------------------

// frame represents a logical span of evaluation logic, i.e. "evaluating resource policies"
type frame struct {
	header string
	hist   []event
}

func (f *frame) emit(msg string, args ...any) {
	next := event{
		eventType: eventTypeMessage,
		message:   format(msg, args...),
	}
	f.hist = append(f.hist, next)
}

func (f *frame) subframe(header string, args ...any) *frame {
	subframe := frame{
		header: format(header, args...),
	}
	next := event{
		eventType: eventTypeSubframe,
		child:     &subframe,
	}
	f.hist = append(f.hist, next)
	return &subframe
}

// -------------------------------------------------------------------------------------------------
// Events
// -------------------------------------------------------------------------------------------------

// eventType is an enum representing the different type of events we may emit
type eventType string

const (
	eventTypeMessage  eventType = "message"
	eventTypeSubframe eventType = "subframe"
)

// event represents an emitted event
type event struct {
	eventType eventType // always present

	message string // only present when eventType == eventTypeMessage
	child   *frame // only present when eventType == eventTypeSubframe
}
