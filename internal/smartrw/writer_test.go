package smartrw

import (
	"fmt"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestNew(t *testing.T) {
	tests := []testlib.TestCase[string, any]{
		{
			Input: "",
		},
		{
			Input:     "/tmp/this/should/not/exist",
			ShouldErr: true,
		},
		{
			Input:     "badproto://test.txt",
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(in string) (any, error) {
		_, err := NewWriter(in)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})
}

func TestSelect(t *testing.T) {
	tests := []testlib.TestCase[string, string]{
		{
			Input: "",
			Want:  "*smartrw.StdoutWriter",
		},
		{
			Input: "/tmp/foo",
			Want:  "*os.File",
		},
		{
			Input: "/tmp/foo.gz",
			Want:  "*smartrw.GzipWriteCloser",
		},
		{
			Input: "file:///tmp/foo",
			Want:  "*os.File",
		},
		{
			Input: "file:///tmp/foo.gz",
			Want:  "*smartrw.GzipWriteCloser",
		},
		{
			Input:     "file:///tmp/should/definitely/not/exist/foo",
			ShouldErr: true,
		},
		{
			Input:     "foo:///tmp/foo",
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(in string) (string, error) {
		w, err := selectWriter(in)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%T", w), nil
	})
}
