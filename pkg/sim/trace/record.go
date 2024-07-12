package trace

// Record is a struct containing an observation, decision, or other trace message emitted during
// evaluation
type Record struct {
	// Message is a human-readable message provided by the evaluation caller
	Message string
	// Frame indicates a friendly name of the current evaluation frame at the time this record was
	// emitted
	Frame string
	// Depth indicates the length of the evaluation stack at the time this record was emitted
	Depth int
	// Attrs indicate any key/value pairs passed up from the evaluation
	Attrs map[string]any
}
