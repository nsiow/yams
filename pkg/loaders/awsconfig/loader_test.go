package awsconfig

import (
	"os"
	"path"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/entities"
)

func TestLoadJson(t *testing.T) {
	tests := []testlib.TestCase[string, entities.Universe]{

		// ---------------------------------------------------------------------------------------------
		// Valid
		// ---------------------------------------------------------------------------------------------

		{
			Name:  "valid_empty_json",
			Input: `../../../testdata/universes/valid_empty.json`,
			Want:  entities.Universe{},
		},

		// ---------------------------------------------------------------------------------------------
		// Invalid
		// ---------------------------------------------------------------------------------------------

	}

	testlib.RunTestSuite(t, tests, func(fp string) (entities.Universe, error) {
		// Load test data
		data, err := os.ReadFile(fp)
		if err != nil {
			t.Fatalf("error while attempting to read test file '%s': %v", fp, err)
		}

		// Call the correct loader based on input type
		l := NewLoader()
		ext := path.Ext(fp)
		switch ext {
		case ".json":
			err = l.LoadJson(data)
		case ".jsonl":
			err = l.LoadJsonl(data)
		default:
			t.Fatalf("unsure how to handle ext '%s'", ext)
		}

		// Handle loading errors; these may be expected
		if err != nil {
			return entities.Universe{}, err
		}
		return *l.Universe(), nil
	})
}
