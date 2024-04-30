package entities

import (
	"reflect"
	"testing"

	"github.com/nsiow/yams/pkg/policy"
)

// TestPolicies validates the merge behavior of the Policies() function
func TestPolicies(t *testing.T) {
	type test struct {
		input Principal
		want  []policy.Policy
	}

	tests := []test{
		{
			input: Principal{
				InlinePolicies:  []policy.Policy{},
				ManagedPolicies: []policy.Policy{},
			},
			want: []policy.Policy{},
		},
		{
			input: Principal{
				InlinePolicies: []policy.Policy{
					{Id: "foo"},
				},
				ManagedPolicies: []policy.Policy{},
			},
			want: []policy.Policy{
				{Id: "foo"},
			},
		},
		{
			input: Principal{
				InlinePolicies: []policy.Policy{},
				ManagedPolicies: []policy.Policy{
					{Id: "foo"},
				},
			},
			want: []policy.Policy{
				{Id: "foo"},
			},
		},
		{
			input: Principal{
				InlinePolicies: []policy.Policy{
					{Id: "foo"},
				},
				ManagedPolicies: []policy.Policy{
					{Id: "bar"},
				},
			},
			want: []policy.Policy{
				{Id: "foo"},
				{Id: "bar"},
			},
		},
	}

	for _, tc := range tests {
		got := tc.input.Policies()
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %#v, got: %#v, for input: %#v", tc.want, got, tc.input)
		}
	}
}

// TestStatements validates the merge behavior of the Statements() function
func TestStatements(t *testing.T) {
	type test struct {
		input Principal
		want  []policy.Statement
	}

	tests := []test{
		{
			input: Principal{
				InlinePolicies:  []policy.Policy{},
				ManagedPolicies: []policy.Policy{},
			},
			want: []policy.Statement(nil),
		},
		{
			input: Principal{
				InlinePolicies: []policy.Policy{
					{
						Statement: policy.StatementBlock{
							Values: []policy.Statement{
								{Sid: "foo"},
							},
						},
					},
				},
				ManagedPolicies: []policy.Policy{},
			},
			want: []policy.Statement{
				{Sid: "foo"},
			},
		},
		{
			input: Principal{
				InlinePolicies: []policy.Policy{
					{
						Statement: policy.StatementBlock{
							Values: []policy.Statement{
								{Sid: "foo"},
							},
						},
					},
				},
				ManagedPolicies: []policy.Policy{
					{
						Statement: policy.StatementBlock{
							Values: []policy.Statement{
								{Sid: "bar"},
							},
						},
					},
				},
			},
			want: []policy.Statement{
				{Sid: "foo"},
				{Sid: "bar"},
			},
		},
	}

	for _, tc := range tests {
		got := tc.input.Statements()
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %#v, got: %#v, for input: %#v", tc.want, got, tc.input)
		}
	}
}
