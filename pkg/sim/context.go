package sim

// Trace provides contextual information around how the
type Trace struct {
	log []string
}

// Add takes the provided message and saves it into our ResultContext
func (r *Trace) Add(message string) {
	r.log = append(r.log, message)
}

// Messages returns all Trace messages received thus far
func (r *Trace) Messages() []string {
	return r.log
}
