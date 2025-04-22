package trace

// Frame represents a logical span of evaluation logic, i.e. "evaluating resource policies"
type Frame struct {
	depth   int
	header  string
	records []string
}

// NewFrame initializes and returns a new frame pointer
func NewFrame(depth int, header string, args ...any) *Frame {
	return &Frame{
		depth:   depth,
		header:  format(header, args...),
		records: nil,
	}
}

// record saves a new message to this frame
func (f *Frame) record(message string) {
	f.records = append(f.records, message)
}

// Records returns all saved messages visible to this frame
func (f *Frame) Records() []string {
	return f.records
}
