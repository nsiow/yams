package trace

type Record struct {
	Message string
	Frame   string
	Depth   int
	Attrs   map[string]any
}
