package smartrw

import (
	"reflect"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
)

func TestReader(t *testing.T) {
	tests := []testlib.TestCase[string, string]{
		{
			Name:  "naked_file",
			Input: "../../testdata/real-world/awsconfig.jsonl",
			Want:  "*os.File",
		},
		{
			Name:  "prefixed_file",
			Input: "file://../../testdata/real-world/awsconfig.jsonl",
			Want:  "*os.File",
		},
		{
			Name:  "basic_gzip",
			Input: "../../testdata/real-world/awsconfig.old.jsonl.gz",
			Want:  "*smartrw.GzipReadCloser",
		},
		{
			Name:  "basic_s3",
			Input: "s3://yams-test-data/data.json",
			Want:  "*smartrw.S3Reader",
		},
		{
			Name:      "bad_s3_path",
			Input:     "s3://yams-test-data",
			ShouldErr: true,
		},
		{
			Name:      "bad_protocol",
			Input:     "bad://some-data-source",
			ShouldErr: true,
		},
		{
			Name:      "bad_gzip",
			Input:     "bad.gz",
			ShouldErr: true,
		},
		{
			// File exists but is not valid gzip - tests gzip error + close path
			Name:      "invalid_gzip_content",
			Input:     "../../testdata/smartrw/invalid.gz",
			ShouldErr: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(in string) (string, error) {
		r, err := NewReader(in)
		if err != nil {
			return "", err
		}

		typeOf := reflect.TypeOf(r.ReadCloser).String()
		return typeOf, nil
	})
}
