package sim

import (
	"testing"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

func TestPrecomputeTargetsSeparatesSameAccountOnlyResources(t *testing.T) {
	actions, err := resolveActions([]string{"sqs:ReceiveMessage"})
	if err != nil {
		t.Fatal(err)
	}
	resources := []*entities.FrozenResource{
		{
			AccountId: "111111111111",
			Arn:       "arn:aws:sqs:us-east-1:111111111111:local-a",
		},
		{
			AccountId: "222222222222",
			Arn:       "arn:aws:sqs:us-east-1:222222222222:local-b",
		},
		{
			AccountId: "222222222222",
			Arn:       "arn:aws:sqs:us-east-1:222222222222:owner-policy",
			Policy: policy.Policy{Statement: []policy.Statement{{
				Effect:   policy.EFFECT_ALLOW,
				Action:   []string{"sqs:ReceiveMessage"},
				Resource: []string{"*"},
				Principal: policy.Principal{AWS: []string{
					"arn:aws:iam::222222222222:root",
				}},
			}}},
		},
		{
			AccountId: "222222222222",
			Arn:       "arn:aws:sqs:us-east-1:222222222222:policy",
			Policy: policy.Policy{Statement: []policy.Statement{{
				Effect:    policy.EFFECT_ALLOW,
				Action:    []string{"sqs:ReceiveMessage"},
				Resource:  []string{"*"},
				Principal: policy.Principal{All: true},
			}}},
		},
	}

	targets := precomputeTargets(actions, resources)
	if len(targets) != 1 {
		t.Fatalf("got %d action targets, want 1", len(targets))
	}
	target := targets[0]
	if target.targetCount != 4 {
		t.Fatalf("got target count %d, want 4", target.targetCount)
	}
	if len(target.resources) != 1 || target.resources[0].index != 3 {
		t.Fatalf("unexpected cross-account candidates: %#v", target.resources)
	}
	if len(target.sameAccountOnly["111111111111"]) != 1 {
		t.Fatalf("unexpected account A targets: %#v", target.sameAccountOnly)
	}
	if len(target.sameAccountOnly["222222222222"]) != 2 {
		t.Fatalf("unexpected account B targets: %#v", target.sameAccountOnly)
	}
}

func TestCachedStatementActionMatchInvalidatesForNewAction(t *testing.T) {
	actions, err := resolveActions([]string{"sqs:ReceiveMessage", "sqs:SendMessage"})
	if err != nil {
		t.Fatal(err)
	}
	statement := &policy.Statement{Action: []string{"sqs:ReceiveMessage"}}
	subject := newSubject(AuthContext{Action: actions[0]}, NewOptions())

	if !evalCachedStatementMatchesAction(&subject, statement) {
		t.Fatal("receive action should match")
	}
	if !evalCachedStatementMatchesAction(&subject, statement) {
		t.Fatal("cached receive action should match")
	}
	subject.auth.Action = actions[1]
	if evalCachedStatementMatchesAction(&subject, statement) {
		t.Fatal("send action should not match")
	}
}

func TestPrincipalPatternMayMatchOtherAccount(t *testing.T) {
	const accountID = "111111111111"
	tests := []struct {
		name    string
		pattern string
		want    bool
	}{
		{"same account ID", accountID, false},
		{"same account root", "arn:aws:iam::" + accountID + ":root", false},
		{"same account role", "arn:aws:iam::" + accountID + ":role/example", false},
		{"other account ID", "222222222222", true},
		{"other account role", "arn:aws:iam::222222222222:role/example", true},
		{"wildcard", "*", true},
		{"unknown", "canonical-id", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := principalPatternMayMatchOtherAccount(test.pattern, accountID)
			if got != test.want {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestResourceMayAllowCrossAccountHandlesEmptyAccount(t *testing.T) {
	if !resourceMayAllowCrossAccount(&entities.FrozenResource{}) {
		t.Fatal("resource without an account must remain a cross-account candidate")
	}
}
