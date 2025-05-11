package sim

import (
	"os"
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/loaders/awsconfig"
)

func buildTestUniverse() (*entities.Universe, error) {
	loader := awsconfig.NewLoader()

	// load resources
	file, err := os.Open("../../testdata/real-world/awsconfig.jsonl")
	if err != nil {
		return nil, err
	}
	err = loader.LoadJsonl(file)
	if err != nil {
		return nil, err
	}

	// load accounts, etc
	file, err = os.Open("../../testdata/real-world/org.jsonl")
	if err != nil {
		return nil, err
	}
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
	sim.options = TestingSimulationOptions
	if err != nil {
		return nil, err
	}
	sim.SetUniverse(uv)

	return sim, nil
}

func TestRealWorldData(t *testing.T) {
	type in struct {
		p string
		a string
		r string
	}

	tests := []testlib.TestCase[in, bool]{
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/RedRole",
				a: "s3:listbucket",
				r: "arn:aws:s3:::yams-magenta",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/RedRole",
				a: "s3:getobject",
				r: "arn:aws:s3:::yams-magenta/object.txt",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/RedRole",
				a: "s3:listbucket",
				r: "arn:aws:s3:::yams-cyan",
			},
			Want: false,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/GreenRole",
				a: "s3:listbucket",
				r: "arn:aws:s3:::yams-green",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/GreenRole",
				a: "s3:getobject",
				r: "arn:aws:s3:::yams-green/secrets.txt",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/GreenRole",
				a: "s3:getobject",
				r: "arn:aws:s3:::yams-green/secrets.txt",
			},
			Want: true,
		},
		{
			Input: in{
				p: "arn:aws:iam::777583092761:role/RedRole",
				a: "s3:getobject",
				r: "arn:aws:s3:::yams-green/secrets.txt",
			},
			Want: false,
		},
	}

	sim, err := buildTestSimulator()
	if err != nil {
		t.Fatalf("error creating simulator for testing: %v", err)
	}

	testlib.RunTestSuite(t, tests, func(i in) (bool, error) {
		result, err := sim.SimulateByArn(i.p, i.a, i.r, nil)
		if err != nil {
			return false, err
		}

		t.Log(result.Trace.String())
		return result.IsAllowed, nil
	})
}
