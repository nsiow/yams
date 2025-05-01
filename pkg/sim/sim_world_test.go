package sim

import (
	"os"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/loaders/awsconfig"
)

func buildTestUniverse() (*entities.Universe, error) {
	file, err := os.Open("testdata/real-world/awsconfig.jsonl")
	if err != nil {
		return nil, err
	}

	loader := awsconfig.NewLoader()
	err = loader.LoadJsonl(file)
	if err != nil {
		return nil, err
	}

	return loader.Universe(), nil
}

func buildTestSimulator() (*Simulator, error) {
	uv, err := buildTestUniverse()
	if err != nil {
		return nil, err
	}

	sim, err := NewSimulator()
	sim.options = *TestingSimulationOptions
	if err != nil {
		return nil, err
	}
	sim.SetUniverse(uv)

	return sim, nil
}

func TestRealWorldData(t *testing.T) {
	type input struct {
		principalArn string
		resourceArn  string
		action       string
	}

	tests := []testlib.TestCase[input, bool]{}

	sim, err := buildTestSimulator()
	if err != nil {
		t.Fatalf("error creating simulator for testing: %v", err)
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		result, err := sim.SimulateByArn(i.principalArn, i.action, i.resourceArn, nil)
		if err != nil {
			return false, err
		}

		t.Log(result.Trace.String())
		return result.IsAllowed, nil
	})
}
