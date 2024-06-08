package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testrunner"
	"github.com/nsiow/yams/pkg/entities"
)

// TestAuthContextKeys validates correct retrieval of Condition keys
func TestAuthContextKeys(t *testing.T) {
	type input struct {
		ac  AuthContext
		key string
	}

	tests := []testrunner.TestCase[input, string]{
		{
			Input: input{
				ac: AuthContext{
					Principal: &entities.Principal{
						Account: "55555",
					},
				},
				key: "aws:PrincipalAccount",
			},
			Want: "55555",
		},
		{
			Input: input{
				ac: AuthContext{
					Resource: &entities.Resource{
						Account: "77777",
					},
				},
				key: "aws:ResourceAccount",
			},
			Want: "77777",
		},
		{
			Input: input{
				ac:  AuthContext{},
				key: "aws:ThisDoesNotExist",
			},
			Want: "",
		},
	}

	testrunner.RunTestSuite(t, tests, func(i input) (string, error) {
		got := i.ac.Key(i.key)
		return got, nil
	})
}
