package sim

import (
	"testing"
)

func TestGate(t *testing.T) {
	type test struct {
		name       string
		numInverts int
		input      bool
		want       bool
	}

	tests := []test{
		{
			name:       "no_inverts_true",
			numInverts: 0,
			input:      true,
			want:       true,
		},
		{
			name:       "no_inverts_false",
			numInverts: 0,
			input:      true,
			want:       true,
		},
		{
			name:       "single_invert_true",
			numInverts: 1,
			input:      true,
			want:       false,
		},
		{
			name:       "single_invert_false",
			numInverts: 1,
			input:      false,
			want:       true,
		},
		{
			name:       "double_invert_true",
			numInverts: 2,
			input:      true,
			want:       true,
		},
		{
			name:       "double_invert_false",
			numInverts: 2,
			input:      false,
			want:       false,
		},
		{
			name:       "high_invert_even_true",
			numInverts: 100,
			input:      true,
			want:       true,
		},
		{
			name:       "high_invert_even_false",
			numInverts: 100,
			input:      false,
			want:       false,
		},
		{
			name:       "high_invert_odd_true",
			numInverts: 151,
			input:      true,
			want:       false,
		},
		{
			name:       "high_invert_odd_false",
			numInverts: 151,
			input:      false,
			want:       true,
		},
	}

	for _, tc := range tests {
		t.Logf("running test case: %s", tc.name)

		// Invert as many times as requested
		g := Gate{}
		for i := 0; i < tc.numInverts; i++ {
			g.Invert()
		}

		// Apply gate
		got := g.Apply(tc.input)
		if got != tc.want {
			t.Fatalf("failed test case '%s': wanted %v got %v", tc.name, got, tc.want)
		}
	}
}
