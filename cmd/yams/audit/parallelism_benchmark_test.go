package audit

import (
	"fmt"
	"testing"

	"github.com/nsiow/yams/pkg/entities"
)

func BenchmarkPrincipalsFromBits(b *testing.B) {
	const principalCount = 154_795
	frozenPrincipals := make([]*entities.FrozenPrincipal, principalCount)
	for i := range frozenPrincipals {
		frozenPrincipals[i] = &entities.FrozenPrincipal{Arn: fmt.Sprintf("principal-%d", i)}
	}

	for _, tc := range []struct {
		name   string
		stride int
	}{
		{name: "sparse_0.2_percent", stride: 500},
		{name: "moderate_10_percent", stride: 10},
		{name: "dense_100_percent", stride: 1},
	} {
		membership := make([]uint64, (principalCount+63)/64)
		for principalIndex := 0; principalIndex < principalCount; principalIndex += tc.stride {
			membership[principalIndex/64] |= 1 << (principalIndex % 64)
		}

		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = principalsFromBits(membership, frozenPrincipals)
			}
		})
	}
}
