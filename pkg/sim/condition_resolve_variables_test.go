package sim

import (
	"testing"

	"github.com/nsiow/yams/pkg/policy"
)

// Mod_ResolveVariables must not mutate the policy in place. The substitution result depends
// on the request context, so a baked-in value would leak across requests.
func TestResolveVariables_DoesNotMutatePolicy(t *testing.T) {
	const placeholder = "${aws:username}"
	stmt := policy.Statement{
		Condition: policy.ConditionBlock{
			"StringEquals": {"aws:Foo": []string{placeholder}},
		},
	}

	// First eval: aws:username = alice, aws:Foo = alice → match.
	subjAlice := newSubject(AuthContext{
		Properties: NewBagFromMap(map[string]string{
			"aws:username": "alice",
			"aws:Foo":      "alice",
		}),
	}, TestingSimulationOptions)
	if !evalStatementMatchesCondition(&subjAlice, &stmt) {
		t.Fatal("first eval (alice) should match")
	}

	// The policy literal must remain ${aws:username}, not "alice".
	got := stmt.Condition["StringEquals"]["aws:Foo"][0]
	if got != placeholder {
		t.Fatalf("policy was mutated by substitution: got %q, want %q", got, placeholder)
	}

	// Second eval: aws:username = bob, aws:Foo = bob → must still match.
	subjBob := newSubject(AuthContext{
		Properties: NewBagFromMap(map[string]string{
			"aws:username": "bob",
			"aws:Foo":      "bob",
		}),
	}, TestingSimulationOptions)
	if !evalStatementMatchesCondition(&subjBob, &stmt) {
		t.Fatal("second eval (bob) should match; substitution likely cached from prior eval")
	}
}
