package trace

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestTrace(t *testing.T) {
	type output struct {
		frames []*Frame
		string []string
	}

	tests := []testlib.TestCase[func(*Trace), []string]{
		{
			Name:  "empty_trace",
			Input: func(t *Trace) {},
			Want: []string{
				"begin: root",
				"end: root",
			},
		},
		{
			Name: "single_message",
			Input: func(t *Trace) {
				t.Observation("foo")
			},
			Want: []string{
				"begin: root",
				"  foo",
				"end: root",
			},
		},
		{
			Name: "multiple_subframes",
			Input: func(t *Trace) {
				t.Observation("foo")
				t.Observation("bar")
				t.Observation("baz")
				t.Push("new thing")
				t.Observation("the")
				t.Observation("quick")
				t.Observation("brown")
				t.Observation("fox")
				t.Push("and another thing")
				t.Observation("lemons")
				t.Pop()
				t.Observation("jumped")
				t.Observation("over")
				t.Pop()
				t.Observation("bao")
			},
			Want: []string{
				"begin: root",
				"  foo",
				"  bar",
				"  baz",
				"  begin: new thing",
				"    the",
				"    quick",
				"    brown",
				"    fox",
				"    begin: and another thing",
				"      lemons",
				"    end: and another thing",
				"    jumped",
				"    over",
				"  end: new thing",
				"  bao",
				"end: root",
			},
		},
	}

	testlib.RunTestSuite(t, tests, func(f func(*Trace)) ([]string, error) {
		trc := New()
		trc.Enable()
		f(trc)

		pr := TestPrinter{}
		trc.Walk(&pr)
		lines := strings.Split(pr.Print(), "\n")

		return lines, nil
	})
}

// TestPrinter is a re-implementation of the [Printer] struct which allows us to more easily test
// and confirm correct functionality without having to refactor tests if we decide to change the
// user-facing formatting of the default [Printer]
type TestPrinter struct {
	messages []string
}

func (p *TestPrinter) Add(s string) {
	p.messages = append(p.messages, s)
}

func (p *TestPrinter) Print() string {
	return strings.Join(p.messages, "\n")
}

func (p *TestPrinter) Indent(fr *Frame) string {
	return strings.Repeat("  ", fr.Depth)
}

func (p *TestPrinter) FrameStart(fr *Frame) {
	p.Add(
		fmt.Sprintf("%sbegin: %s", p.Indent(fr), fr.Header),
	)
}

func (p *TestPrinter) Message(fr *Frame, msg string) {
	p.Add(
		fmt.Sprintf("%s  %s", p.Indent(fr), msg),
	)
}

func (p *TestPrinter) FrameEnd(fr *Frame) {
	p.Add(
		fmt.Sprintf("%send: %s", p.Indent(fr), fr.Header),
	)
}
