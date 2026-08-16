package sim

import (
	"context"
	"fmt"
	"testing"

	"github.com/nsiow/yams/pkg/entities"
	"github.com/nsiow/yams/pkg/policy"
)

func benchmarkProductInputs(allow bool) (
	[]*entities.FrozenPrincipal,
	[]string,
	[]*entities.FrozenResource,
) {
	const (
		principalCount = 2048
		resourceCount  = 64
	)

	principalAccount := "111111111111"
	resourceAccount := "222222222222"
	if allow {
		resourceAccount = principalAccount
	}

	principals := make([]*entities.FrozenPrincipal, principalCount)
	for i := range principals {
		arn := fmt.Sprintf("arn:aws:iam::%s:role/principal-%d", principalAccount, i)
		principals[i] = &entities.FrozenPrincipal{
			Type:        "AWS::IAM::Role",
			AccountId:   principalAccount,
			Arn:         arn,
			ArnSegments: entities.SplitArn(arn),
		}
		if allow {
			principals[i].InlinePolicies = []policy.Policy{{
				Statement: []policy.Statement{{
					Effect:   policy.EFFECT_ALLOW,
					Action:   []string{"sqs:*"},
					Resource: []string{"*"},
				}},
			}}
		}
	}

	resources := make([]*entities.FrozenResource, resourceCount)
	for i := range resources {
		arn := fmt.Sprintf("arn:aws:sqs:us-east-1:%s:queue-%d", resourceAccount, i)
		resources[i] = &entities.FrozenResource{
			Type:        "AWS::SQS::Queue",
			AccountId:   resourceAccount,
			Arn:         arn,
			ArnSegments: entities.SplitArn(arn),
		}
	}

	return principals, []string{
		"sqs:ReceiveMessage",
		"sqs:SendMessage",
		"sqs:DeleteMessage",
		"sqs:PurgeQueue",
	}, resources
}

func BenchmarkProductFrozen(b *testing.B) {
	for _, tc := range []struct {
		name  string
		allow bool
	}{
		{name: "cross_account_deny"},
		{name: "same_account_allow", allow: true},
	} {
		principals, actions, resources := benchmarkProductInputs(tc.allow)
		simulator, err := NewSimulator()
		if err != nil {
			b.Fatal(err)
		}
		opts := NewOptions()

		b.Run(tc.name+"/indexed", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				err := simulator.ProductFrozenIndexed(
					context.Background(),
					principals,
					actions,
					resources,
					opts,
					func(IndexedAccess) error { return nil },
				)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(tc.name+"/streaming", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				err := simulator.ProductFrozenStreaming(
					principals,
					actions,
					resources,
					opts,
					func(AccessTuple) {},
				)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
