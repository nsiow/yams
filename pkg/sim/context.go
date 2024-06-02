package sim

// ResultContext provides contextual information around how the
type ResultContext struct {
	log []string
}

// Add takes the provided message and saves it into our ResultContext
func (r *ResultContext) Add(message string) {
	r.log = append(r.log, message)
}
