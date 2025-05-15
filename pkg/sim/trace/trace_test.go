package trace

import (
	"strings"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestTrace(t *testing.T) {
	type output struct {
		Explain []string
		Trace   []string
		Print   string
	}

	tests := []testlib.TestCase[func(*Trace), output]{
		{
			Name:  "empty_trace",
			Input: func(t *Trace) {},
			Want: output{
				Trace: []string{
					"begin: root",
					"end: root",
				},
				Print: strings.Join([]string{
					"begin: root",
					"end: root",
				}, "\n"),
			},
		},
		{
			Name: "single_message",
			Input: func(t *Trace) {
				t.Enable()
				t.Log("foo")
			},
			Want: output{
				Trace: []string{
					"begin: root",
					"  foo",
					"end: root",
				},
				Print: strings.Join([]string{
					"begin: root",
					"  foo",
					"end: root",
				}, "\n"),
			},
		},
		{
			Name: "enable_disable",
			Input: func(t *Trace) {
				t.Enable()
				t.Log("foo")
				t.Disable()
				t.Push("new thing")
				t.Allowed("some decision evaluated to true")
				t.Denied("some decision evaluated to false")
				t.Pop()
				t.Log("bar")
				t.Enable()
				t.Log("baz")
			},
			Want: output{
				Trace: []string{
					"begin: root",
					"  foo",
					"  baz",
					"end: root",
				},
				Print: strings.Join([]string{
					"begin: root",
					"  foo",
					"  baz",
					"end: root",
				}, "\n"),
			},
		},
		{
			Name: "multiple_subframes",
			Input: func(t *Trace) {
				t.Enable()
				t.Log("foo")
				t.Log("bar")
				t.Log("baz")
				t.Push("new thing")
				t.Log("the")
				t.Log("quick")
				t.Log("brown %s", "fox")
				t.Push("and another thing")
				t.Allowed("yes")
				t.Log("lemons")
				t.Pop()
				t.Denied("no")
				t.Log("jumped")
				t.Log("over")
				t.Pop()
				t.Log("bao")
			},
			Want: output{
				Explain: []string{
					"the",
					"quick",
					"brown fox",
					"yes",
					"lemons",
					"no",
					"jumped",
					"over",
				},
				Trace: []string{
					"begin: root",
					"  foo",
					"  bar",
					"  baz",
					"  (deny) begin: new thing",
					"    the",
					"    quick",
					"    brown fox",
					"    (allow) begin: and another thing",
					"      yes",
					"      lemons",
					"    end: and another thing",
					"    no",
					"    jumped",
					"    over",
					"  end: new thing",
					"  bao",
					"end: root",
				},
				Print: strings.Join([]string{
					"begin: root",
					"  foo",
					"  bar",
					"  baz",
					"  (deny) begin: new thing",
					"    the",
					"    quick",
					"    brown fox",
					"    (allow) begin: and another thing",
					"      yes",
					"      lemons",
					"    end: and another thing",
					"    no",
					"    jumped",
					"    over",
					"  end: new thing",
					"  bao",
					"end: root",
				}, "\n"),
			},
		},
	}

	testlib.RunTestSuite(t, tests, func(f func(*Trace)) (output, error) {
		trc := New()
		f(&trc)

		return output{
			Explain: trc.Explain(),
			Trace:   trc.Trace(),
			Print:   trc.Print(),
		}, nil
	})
}

func TestPanicNoStack(t *testing.T) {
	defer testlib.AssertPanicWithText(t,
		"attempt to look up current frame for empty stack")

	trc := Trace{}
	trc.Enable()
	trc.curr()
}

func TestPanicPopRoot(t *testing.T) {
	defer testlib.AssertPanicWithText(t,
		"attempt to pop root frame from trace stack")

	trc := New()
	trc.Enable()
	trc.Pop()
}

func TestPanicEmptyWalk(t *testing.T) {
	defer testlib.AssertPanicWithText(t,
		"trace somehow has empty stack")

	pr := Printer{}
	trc := Trace{}
	trc.Enable()
	trc.Walk(&pr)
}

func TestPanicBadEventWalk(t *testing.T) {
	defer testlib.AssertPanicWithText(t,
		"unexpected event type: weird")

	pr := Printer{}
	trc := New()
	trc.Enable()
	trc.Log("foo")

	// mess up the event type
	trc.stack[len(trc.stack)-1].hist[0].eventType = "weird"

	trc.Walk(&pr)
}

// Just tests required to hit coverage targets for empty/unused function bodies
func TestExtra(t *testing.T) {
	e := Explainer{}
	e.FrameStart(nil)
	e.FrameEnd(nil)
}
