package sim

// Gate is a helper struct implementing a NOT gate for easier handling of policy inversions
type Gate struct {
	inverted bool
}

// Invert instructs Gate to invert whatever its current inversion instruction is
func (g *Gate) Invert() {
	g.inverted = !g.inverted
}

// Apply instructs Gate to use its current inversion instruction on the supplied boolean
func (g *Gate) Apply(b bool) bool {
	if g.inverted {
		return !b
	}

	return b
}
