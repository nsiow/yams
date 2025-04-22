package trace

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestTrace(t *testing.T) {
	tests := []testlib.TestCase[func(*Trace), []*frame]{
		{
			Name:  "empty_trace",
			Input: func(t *Trace) {},
			Want: []*frame{
				{header: "root"},
			},
		},
		// {
		// 	Name: "single_message",
		// 	Input: func(t *Trace) {
		// 		t.Observation("foo")
		// 	},
		// 	Want: []*frame{
		// 		{
		// 			header: "root",
		// 			hist:   []event{},
		// 		},
		// 	},
		// },
	}

	testlib.RunTestSuite(t, tests, func(f func(*Trace)) ([]*frame, error) {
		trc := New()
		f(trc)
		return trc.stack, nil
	})
}
