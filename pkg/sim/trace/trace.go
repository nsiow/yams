package trace

type Trace struct {
	minLevel int
	stack    []string
	attrs    []map[string]any
	buf      []Record
}

func New() *Trace {
	// TODO(nsiow) maybe preallocate space here
	trc := &Trace{}
	trc.Push("root")
	return trc
}

func (t *Trace) Push(frame string) {
	t.stack = append(t.stack, frame)
	t.attrs = append(t.attrs, make(map[string]any))
}

func (t *Trace) Pop() {
	if len(t.stack) > 1 {
		t.stack = t.stack[:len(t.stack)-1]
	}
	if len(t.attrs) > 1 {
		t.attrs = t.attrs[:len(t.attrs)-1]
	}
}

func (t *Trace) Attr(k string, v any) {
	t.attrs[len(t.attrs)-1][k] = v
}

func (t *Trace) History() []Record {
	return t.buf
}

func (t *Trace) SetLevel(l Level) {
	t.minLevel = l
}

func (t *Trace) Comparison(msg string) {
	if t.minLevel <= LEVEL_COMPARISON {
		t.save(msg)
	}
}

func (t *Trace) Decision(msg string) {
	if t.minLevel <= LEVEL_DECISION {
		t.save(msg)
	}
}

func (t *Trace) save(msg string) {
	attrcopy := make(map[string]any)
	for k, v := range t.attrs[len(t.attrs)-1] {
		attrcopy[k] = v
	}

	r := Record{
		Message: msg,
		Attrs:   attrcopy,
		Frame:   t.stack[len(t.stack)-1],
		Depth:   len(t.stack) - 1,
	}
	t.buf = append(t.buf, r)
}
