package trace

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestTrace(t *testing.T) {
	tests := []testlib.TestCase[func(*Trace), []*Frame]{
		{
			Input: func(t *Trace) {

			},
			Want: []*Frame{},
		},
	}

	testlib.RunTestSuite(t, tests, func(f func(*Trace)) ([]*Frame, error) {
		trc := New()
		f(trc)
		return trc.History(), nil
	})
}
