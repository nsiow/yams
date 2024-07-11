package trace

import (
	"reflect"
	"testing"
)

// TestTraceLevel validates the behavior Trace levels
func TestTraceLevel(t *testing.T) {
	// Create a new trace
	trc := New()
	trc.SetLevel(LEVEL_COMPARISON)

	// Add COMPARISON, should be entered
	trc.Comparison("test")
	want := 1
	got := trc.History()
	if want != len(trc.History()) {
		t.Fatalf("wanted (%+v), got (%+v)", want, got)
	}

	// Add DECISION, should be entered
	trc.Decision("test")
	want = 2
	got = trc.History()
	if want != len(trc.History()) {
		t.Fatalf("wanted (%+v), got (%+v)", want, got)
	}

	// Update level
	trc.SetLevel(LEVEL_DECISION)

	// Add COMPARISON, should NOT be entered
	trc.Comparison("test")
	want = 2
	got = trc.History()
	if want != len(trc.History()) {
		t.Fatalf("wanted (%+v), got (%+v)", want, got)
	}

	// Add DECISION, should be entered
	trc.Decision("test")
	want = 3
	got = trc.History()
	if want != len(trc.History()) {
		t.Fatalf("wanted (%+v), got (%+v)", want, got)
	}
}

// TestTraceSingle validates the behavior of a basic simulation Trace
func TestTraceSingle(t *testing.T) {
	// Create a new trace
	trc := New()

	// Add single item, confirm depth + content
	trc.Attr("foo", "bar")
	trc.Comparison("hello world")

	// Compare history to expected
	want := []Record{
		{
			Message: "hello world",
			Frame:   "root",
			Depth:   0,
			Attrs: map[string]any{
				"foo": "bar",
			},
		},
	}
	got := trc.History()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("wanted (%+v), got (%+v)", want, got)
	}
}

// TestTraceMany validates the behavior of a nested simulation Trace
func TestTraceMany(t *testing.T) {
	// Create a new trace
	trc := New()

	// Add single attr, item
	trc.Attr("foo", "bar")
	trc.Comparison("hello world")

	// Create new frame; add two attrs, one item
	trc.Push("first")
	trc.Attr("water", "melon")
	trc.Attr("sweet", "potato")
	trc.Comparison("hello world 2")

	// Create final frame; add single attr, item
	trc.Push("second")
	trc.Attr("egg", "yolk")
	trc.Comparison("hello world 3")

	// Pop frame; add single attr, item
	trc.Pop()
	trc.Attr("olive", "oil")
	trc.Comparison("hello world 4")

	// Pop frame; add single item
	trc.Pop()
	trc.Attr("chicken", "soup")
	trc.Comparison("hello world 5")

	// Pop several times, confirm we still have root frame
	trc.Pop()
	trc.Pop()
	trc.Pop()
	trc.Attr("artichoke", "heart")
	trc.Comparison("hello world 6")

	// Compare history to expected
	want := []Record{
		{
			Message: "hello world",
			Frame:   "root",
			Depth:   0,
			Attrs: map[string]any{
				"foo": "bar",
			},
		},
		{
			Message: "hello world 2",
			Frame:   "first",
			Depth:   1,
			Attrs: map[string]any{
				"water": "melon",
				"sweet": "potato",
			},
		},
		{
			Message: "hello world 3",
			Frame:   "second",
			Depth:   2,
			Attrs: map[string]any{
				"egg": "yolk",
			},
		},
		{
			Message: "hello world 4",
			Frame:   "first",
			Depth:   1,
			Attrs: map[string]any{
				"water": "melon",
				"sweet": "potato",
				"olive": "oil",
			},
		},
		{
			Message: "hello world 5",
			Frame:   "root",
			Depth:   0,
			Attrs: map[string]any{
				"foo":     "bar",
				"chicken": "soup",
			},
		},
		{
			Message: "hello world 6",
			Frame:   "root",
			Depth:   0,
			Attrs: map[string]any{
				"foo":       "bar",
				"chicken":   "soup",
				"artichoke": "heart",
			},
		},
	}
	got := trc.History()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("wanted\n%+v\n\ngot\n%+v", want, got)
	}
}
