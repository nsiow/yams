package trace

// Level defines logging-esque levels of different trace statements to record
type Level = int

const (
	// Record all observations
	LEVEL_OBSERVATION = 0

	// Record all decisions
	LEVEL_DECISION = 1
)
