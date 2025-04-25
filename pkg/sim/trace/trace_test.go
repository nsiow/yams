package trace

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestTrace(t *testing.T) {
	type output struct {
		frames []Frame
		string []string
	}

	tests := []testlib.TestCase[func(*Trace), output]{
		{
			Name:  "empty_trace",
			Input: func(t *Trace) {},
			Want: output{
				frames: []Frame{
					{
						Header: "root",
					},
				},
				string: []string{
					"begin: root",
					"end: root",
				},
			},
		},
		{
			Name: "single_message",
			Input: func(t *Trace) {
				t.Observation("foo")
			},
			Want: output{
				frames: []Frame{
					{
						Header: "root",
						hist: []event{
							{
								eventType: eventTypeMessage,
								message:   "foo",
							},
						},
					},
				},
				string: []string{
					"begin: root",
					"  foo",
					"end: root",
				},
			},
		},
		// {
		// 	Name: "single_message",
		// 	Input: func(t *Trace) {
		// 		t.Observation("foo")
		// 	},
		// 	Want: output{
		// 		frames: []*frame{
		// 			{
		// 				header: "root",
		// 				hist: []event{
		// 					{
		// 						eventType: eventTypeMessage,
		// 						message:   "foo",
		// 					},
		// 				},
		// 			},
		// 		},
		// 		string: "",
		// 	},
		// },
	}

	testlib.RunTestSuite(t, tests, func(f func(*Trace)) (output, error) {
		trc := New()
		trc.Enable()
		f(trc)

		pr := TestPrinter{}
		trc.Walk(&pr)

		return output{
			frames: trc.stack,
			string: strings.Split(pr.Print(), "\n"),
		}, nil
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

func (p *TestPrinter) Indent(fr Frame) string {
	return strings.Repeat("  ", fr.Depth)
}

func (p *TestPrinter) FrameStart(fr Frame) {
	p.Add(
		fmt.Sprintf("%sbegin: %s", p.Indent(fr), fr.Header),
	)
}

func (p *TestPrinter) Message(fr Frame, msg string) {
	p.Add(
		fmt.Sprintf("%s  %s", p.Indent(fr), msg),
	)
}

func (p *TestPrinter) FrameEnd(fr Frame) {
	p.Add(
		fmt.Sprintf("%send: %s", p.Indent(fr), fr.Header),
	)
}
