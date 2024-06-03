package sim

import (
	"testing"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

// TestStatementMatchesAction checks action-matching logic for statements
func TestStatementMatchesAction(t *testing.T) {
	type test struct {
		name  string
		event Event
		stmt  policy.Statement
		want  bool
		err   bool
	}

	tests := []test{
		// Action
		// {
		// 	name:  "simple_wildcard",
		// 	event: Event{Action: "s3:getobject"},
		// 	stmt:  policy.Statement{Action: []string{"*"}},
		// 	want:  true,
		// },
		// {
		// 	name:  "simple_direct_match",
		// 	event: Event{Action: "s2:getobject"},
		// 	stmt:  policy.Statement{Action: []string{"s2:getobject"}},
		// 	want:  true,
		// },
		{
			name:  "other_action",
			event: Event{Action: "s3:putobject"},
			stmt:  policy.Statement{Action: []string{"s3:getobject"}},
			want:  false,
		},
		{
			name:  "two_actions",
			event: Event{Action: "s3:getobject"},
			stmt:  policy.Statement{Action: []string{"s3:putobject", "s3:getobject"}},
			want:  true,
		},
		{
			name:  "diff_casing",
			event: Event{Action: "s3:gEtObJeCt"},
			stmt:  policy.Statement{Action: []string{"s3:putobject", "s3:getobject"}},
			want:  true,
		},

		// NotAction
		{
			name:  "notaction_simple_wildcard",
			event: Event{Action: "s3:getobject"},
			stmt:  policy.Statement{NotAction: []string{"*"}},
			want:  false,
		},
		{
			name:  "notaction_simple_direct_match",
			event: Event{Action: "s3:getobject"},
			stmt:  policy.Statement{NotAction: []string{"s3:getobject"}},
			want:  false,
		},
		{
			name:  "notaction_other_action",
			event: Event{Action: "sqs:sendmessage"},
			stmt:  policy.Statement{NotAction: []string{"s3:getobject"}},
			want:  true,
		},
		{
			name:  "notaction_two_actions",
			event: Event{Action: "s3:getobject"},
			stmt:  policy.Statement{NotAction: []string{"s3:putobject", "s3:getobject"}},
			want:  false,
		},
		{
			name:  "notaction_diff_casing",
			event: Event{Action: "s3:gEtObJeCt"},
			stmt:  policy.Statement{NotAction: []string{"s3:putobject", "s3:getobject"}},
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		opts := SimOptions{}
		trc := Trace{}
		got, err := evalStatementMatchesAction(&opts, &tc.event, &trc, &tc.stmt)
		if err != nil {
			if tc.err {
				t.Logf("observed expected error: %v", err)
			} else {
				t.Fatalf("observed unexpected error: %v", err)
			}
		}

		if got != tc.want {
			t.Fatalf("failed test case: '%s', wanted %v got %v", tc.name, tc.want, got)
		}
	}
}

// TestStatementMatchesPrincipal checks principal-matching logic for statements
func TestStatementMatchesPrincipal(t *testing.T) {
	type test struct {
		name  string
		event Event
		stmt  policy.Statement
		want  bool
		err   bool
	}

	tests := []test{
		// Principal
		{
			name:  "simple_wildcard",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::55555:role/somerole"}},
			stmt:  policy.Statement{Principal: policy.Principal{AWS: []string{"*"}}},
			want:  true,
		},
		{
			name:  "simple_direct_match",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::55555:role/somerole"}},
			stmt:  policy.Statement{Principal: policy.Principal{AWS: []string{"arn:aws:iam::55555:role/somerole"}}},
			want:  true,
		},
		{
			name:  "other_principal",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::55555:role/somerole"}},
			stmt:  policy.Statement{Principal: policy.Principal{AWS: []string{"arn:aws:iam::55555:role/somerandomrole"}}},
			want:  false,
		},
		{
			name:  "two_principals",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::55555:role/secondrole"}},
			stmt: policy.Statement{Principal: policy.Principal{AWS: []string{
				"arn:aws:iam::55555:role/firstrole",
				"arn:aws:iam::55555:role/secondrole"}}},
			want: true,
		},
		{
			name:  "other_service",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::55555:role/somerole"}},
			stmt:  policy.Statement{Principal: policy.Principal{Federated: []string{"*"}}},
			want:  false,
		},

		// NotPrincipal
		{
			name:  "notprincipal_simple_wildcard",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::55555:role/somerole"}},
			stmt:  policy.Statement{NotPrincipal: policy.Principal{AWS: []string{"*"}}},
			want:  false,
		},
		{
			name:  "notprincipal_simple_direct_match",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::55555:role/somerole"}},
			stmt:  policy.Statement{NotPrincipal: policy.Principal{AWS: []string{"arn:aws:iam::55555:role/somerole"}}},
			want:  false,
		},
		{
			name:  "notprincipal_other_principal",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::55555:role/somerole"}},
			stmt:  policy.Statement{NotPrincipal: policy.Principal{AWS: []string{"arn:aws:iam::55555:role/somerandomrole"}}},
			want:  true,
		},
		{
			name:  "notprincipal_two_principals",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::55555:role/secondrole"}},
			stmt: policy.Statement{NotPrincipal: policy.Principal{AWS: []string{
				"arn:aws:iam::55555:role/firstrole",
				"arn:aws:iam::55555:role/secondrole"}}},
			want: false,
		},
		{
			name:  "notprincipal_other_service",
			event: Event{Principal: &entities.Principal{Arn: "arn:aws:iam::55555:role/somerole"}},
			stmt:  policy.Statement{NotPrincipal: policy.Principal{Federated: []string{"*"}}},
			want:  true,
		},
	}

	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		opts := SimOptions{}
		trc := Trace{}
		got, err := evalStatementMatchesPrincipal(&opts, &tc.event, &trc, &tc.stmt)
		if err != nil {
			if tc.err {
				t.Logf("observed expected error: %v", err)
			} else {
				t.Fatalf("observed unexpected error: %v", err)
			}
		}

		if got != tc.want {
			t.Fatalf("failed test case: '%s', wanted %v got %v", tc.name, tc.want, got)
		}
	}
}

// TestStatementMatchesResource checks resource-matching logic for statements
func TestStatementMatchesResource(t *testing.T) {
	type test struct {
		name  string
		event Event
		stmt  policy.Statement
		want  bool
		err   bool
	}

	tests := []test{
		// Resource
		{
			name:  "simple_wildcard",
			event: Event{Resource: &entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
			stmt:  policy.Statement{Resource: []string{"*"}},
			want:  true,
		},
		{
			name:  "simple_direct_match",
			event: Event{Resource: &entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
			stmt:  policy.Statement{Resource: []string{"arn:aws:s3:::somebucket"}},
			want:  true,
		},
		{
			name:  "other_resource",
			event: Event{Resource: &entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
			stmt:  policy.Statement{Resource: []string{"arn:aws:s3:::adifferentbucket"}},
			want:  false,
		},
		{
			name:  "two_resources",
			event: Event{Resource: &entities.Resource{Arn: "arn:aws:s3:::secondbucket"}},
			stmt: policy.Statement{Resource: []string{
				"arn:aws:s3:::firstbucket",
				"arn:aws:s3:::secondbucket"}},
			want: true,
		},

		// NotResource
		{
			name:  "notresource_simple_wildcard",
			event: Event{Resource: &entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
			stmt:  policy.Statement{NotResource: []string{"*"}},
			want:  false,
		},
		{
			name:  "notresource_simple_direct_match",
			event: Event{Resource: &entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
			stmt:  policy.Statement{NotResource: []string{"arn:aws:s3:::somebucket"}},
			want:  false,
		},
		{
			name:  "notresource_other_resource",
			event: Event{Resource: &entities.Resource{Arn: "arn:aws:s3:::somebucket"}},
			stmt:  policy.Statement{NotResource: []string{"arn:aws:s3:::adifferentbucket"}},
			want:  true,
		},
		{
			name:  "notresource_two_resources",
			event: Event{Resource: &entities.Resource{Arn: "arn:aws:s3:::secondbucket"}},
			stmt: policy.Statement{NotResource: []string{
				"arn:aws:s3:::firstbucket",
				"arn:aws:s3:::secondbucket"}},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		opts := SimOptions{}
		trc := Trace{}
		got, err := evalStatementMatchesResource(&opts, &tc.event, &trc, &tc.stmt)
		if err != nil {
			if tc.err {
				t.Logf("observed expected error: %v", err)
			} else {
				t.Fatalf("observed unexpected error: %v", err)
			}
		}

		if got != tc.want {
			t.Fatalf("failed test case: '%s', wanted %v got %v", tc.name, tc.want, got)
		}
	}
}

// TestStatementMatchesCondition checks condition-matching logic for statements
func TestStatementMatchesCondition(t *testing.T) {
	// FIXME(nsiow)
	opts := SimOptions{}
	evt := Event{}
	trc := Trace{}
	stmt := policy.Statement{}
	evalStatementMatchesCondition(&opts, &evt, &trc, &stmt)
}

// TestEvalIsSameAccount checks same vs x-account checking behavior
func TestEvalIsSameAccount(t *testing.T) {
	type test struct {
		principal entities.Principal
		resource  entities.Resource
		want      bool
	}

	tests := []test{
		{
			principal: entities.Principal{Account: "55555"},
			resource:  entities.Resource{Account: "55555"},
			want:      true,
		},
		{
			principal: entities.Principal{Account: "55555"},
			resource:  entities.Resource{Account: "12345"},
			want:      false,
		},
	}

	for _, tc := range tests {
		got := evalIsSameAccount(&tc.principal, &tc.resource)
		if got != tc.want {
			t.Fatalf("failed, wanted %v got %v for test case: %v", tc.want, got, tc)
		}
	}
}
