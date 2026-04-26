package sim

import (
	"testing"

	"github.com/nsiow/yams/internal/testlib"
	"github.com/nsiow/yams/pkg/policy"
)

// AWS accepts both CIDR (e.g. 10.0.0.0/8) and bare IP (e.g. 192.0.2.1) on the policy side
// of IpAddress / NotIpAddress. Bare IPs are treated as host routes.
func TestIpAddress_BareAndCIDR(t *testing.T) {
	tests := []testlib.TestCase[input, bool]{
		{
			Name: "bare_ipv4_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIp": "192.0.2.1",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"IpAddress": {"aws:SourceIp": []string{"192.0.2.1"}},
					},
				},
			},
			Want: true,
		},
		{
			Name: "bare_ipv4_no_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIp": "192.0.2.2",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"IpAddress": {"aws:SourceIp": []string{"192.0.2.1"}},
					},
				},
			},
			Want: false,
		},
		{
			Name: "cidr_still_works",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIp": "10.1.2.3",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"IpAddress": {"aws:SourceIp": []string{"10.0.0.0/8"}},
					},
				},
			},
			Want: true,
		},
		{
			Name: "bare_ipv6_match",
			Input: input{
				ac: AuthContext{
					Properties: NewBagFromMap(map[string]string{
						"aws:SourceIp": "2001:db8::1",
					}),
				},
				stmt: policy.Statement{
					Condition: policy.ConditionBlock{
						"IpAddress": {"aws:SourceIp": []string{"2001:db8::1"}},
					},
				},
			},
			Want: true,
		},
	}

	testlib.RunTestSuite(t, tests, func(i input) (bool, error) {
		subj := newSubject(i.ac, TestingSimulationOptions)
		return evalStatementMatchesCondition(&subj, &i.stmt), nil
	})
}
