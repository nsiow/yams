package sim

import (
	"github.com/nsiow/yams/pkg/sim/trace"
)

// evalResourceAccess calculates the Resource-side access with regard to the specified Principal
func evalResourceAccess(
	trc *trace.Trace,
	opt *Options,
	ac AuthContext) (EffectSet, error) {

	trc.Push("evaluating resource policies")
	defer trc.Pop()

	// Specify the statement evaluation funcs we will consider for Principal access
	funcs := []evalFunction{
		evalStatementMatchesAction,
		evalStatementMatchesPrincipal,
		evalStatementMatchesCondition,
	}

	// Iterate over resource policy statements to evaluate access
	return evalPolicy(trc, opt, ac, ac.Resource.Policy, funcs)
}
